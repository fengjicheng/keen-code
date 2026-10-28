package llm

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"reflect"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"strings"

	"github.com/mochow13/keen-code/internal/llm/core"
	"testing"

	"github.com/mochow13/keen-code/internal/tools"
)

type compactionTestClient struct {
	request  []core.Message
	registry *tools.Registry
	opts     core.StreamOptions
	events   []core.StreamEvent
}

func (c *compactionTestClient) StreamChat(_ context.Context, messages []core.Message, registry *tools.Registry, opts ...core.StreamOptions) (<-chan core.StreamEvent, error) {
	c.request = core.CloneMessages(messages)
	c.registry = registry
	c.opts = streamOptions(opts)
	ch := make(chan core.StreamEvent, len(c.events))
	for _, event := range c.events {
		ch <- event
	}
	close(ch)
	return ch, nil
}
func (*compactionTestClient) Reset() {}

func TestAutoCompactBuildsPrivateReplacement(t *testing.T) {
	client := &compactionTestClient{events: []core.StreamEvent{{Type: core.StreamEventTypeChunk, Content: "## Goal\nContinue work"}}}
	history := []core.Message{{Role: core.RoleSystem, Content: "agent prompt"}, {Role: core.RoleUser, Content: "implement this exactly"}}
	registry := tools.NewRegistry()
	replacement, _, err := AutoCompact(context.Background(), client, history, registry, "session")
	if err != nil {
		t.Fatal(err)
	}
	if !client.opts.OneShot || !client.opts.DisableAutoCompaction || !client.opts.DisableToolCalls || client.opts.SessionID != "session" {
		t.Fatalf("unexpected nested options: %#v", client.opts)
	}
	if client.registry != registry {
		t.Fatal("automatic compaction did not retain the tool registry")
	}
	if len(client.request) != len(history)+1 || client.request[0].Role != core.RoleSystem || client.request[0].Content != "agent prompt" {
		t.Fatalf("history prefix was not preserved: %#v", client.request)
	}
	last := client.request[len(client.request)-1]
	if last.Role != core.RoleUser || !strings.Contains(last.Content, "context compaction request") {
		t.Fatalf("compaction instruction was not appended as final user message: %#v", client.request)
	}
	if len(replacement) != 2 || replacement[0].Role != core.RoleSystem || replacement[0].Content != "agent prompt" || replacement[1].Role != core.RoleUser || !strings.Contains(replacement[1].Content, "implement this exactly") {
		t.Fatalf("invalid replacement: %#v", replacement)
	}
	replacement[0].Content = "changed"
	if history[0].Content != "agent prompt" {
		t.Fatal("replacement system message aliases original history")
	}
}

func TestAutoCompactIsTransactional(t *testing.T) {
	history := []core.Message{{Role: core.RoleUser, Content: "task"}}
	client := &compactionTestClient{}
	_, _, err := AutoCompact(context.Background(), client, history, nil, "")
	if err == nil || !strings.Contains(err.Error(), "empty summary") {
		t.Fatalf("expected empty summary error, got %v", err)
	}
	if history[0].Content != "task" {
		t.Fatal("history mutated")
	}
}
func TestAutoCompactUsesOnlyFinalAssistantRunAfterToolActivity(t *testing.T) {
	client := &compactionTestClient{events: []core.StreamEvent{
		{Type: core.StreamEventTypeChunk, Content: "preamble"},
		{Type: core.StreamEventTypeToolStart},
		{Type: core.StreamEventTypeToolEnd},
		{Type: core.StreamEventTypeChunk, Content: "final summary"},
	}}
	replacement, _, err := AutoCompact(context.Background(), client, []core.Message{{Role: core.RoleUser, Content: "task"}}, tools.NewRegistry(), "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(replacement[0].Content, "preamble") || !strings.Contains(replacement[0].Content, "final summary") {
		t.Fatalf("unexpected replacement: %#v", replacement)
	}
}

func TestAutoCompactRejectsIncompleteStream(t *testing.T) {
	client := &compactionTestClient{events: []core.StreamEvent{
		{Type: core.StreamEventTypeChunk, Content: "partial"},
		{Type: core.StreamEventTypeIncomplete},
	}}
	_, _, err := AutoCompact(context.Background(), client, []core.Message{{Role: core.RoleUser, Content: "task"}}, nil, "")
	if err == nil || !strings.Contains(err.Error(), "stream incomplete") {
		t.Fatalf("expected incomplete-stream error, got %v", err)
	}
}

func TestContextBudgetAndAutoCompactionThreshold(t *testing.T) {
	if got := core.ContextInputBudget(100000); got != 95000 {
		t.Fatalf("budget = %d, want 95000", got)
	}
	if !core.ShouldAutoCompact(85500, 95000) || core.ShouldAutoCompact(85499, 95000) {
		t.Fatal("unexpected 90% trigger boundary")
	}
}

func TestRemainingClientsPreserveInputHistory(t *testing.T) {
	for _, provider := range []string{"anthropic", "bedrock", "genkit", "responses"} {
		for _, mode := range []string{"disabled", "success", "failure", "cancelled"} {
			t.Run(provider+"/"+mode, func(t *testing.T) {
				backing := append(openAIHistoryFixture(), core.Message{Role: core.RoleUser, Content: "sentinel"})
				messages := backing[:len(backing)-1]
				before := core.CloneMessages(backing)
				calls := 0
				var compactionInput string
				next := func(request any) (bool, string, error) {
					calls++
					if calls == 1 {
						return true, "", nil
					}
					if calls == 2 && mode != "disabled" {
						encoded, err := json.Marshal(request)
						if err != nil {
							t.Error(err)
						}
						compactionInput = string(encoded)
						switch mode {
						case "failure":
							return false, "", errors.New("compaction failed")
						case "cancelled":
							return false, "", context.Canceled
						}
						return false, "compacted summary", nil
					}
					return false, "done", nil
				}
				var client LLMClient
				switch provider {
				case "anthropic":
					c := &AnthropicClient{model: "test", maxRetries: 1, contextWindowTokenCount: 1}
					c.streamImpl = func(_ context.Context, params anthropic.MessageNewParams, _ ...anthropicoption.RequestOption) anthropicStream {
						tool, text, err := next(params)
						if tool {
							return &mockAnthropicStream{events: []anthropic.MessageStreamEventUnion{
								makeToolUseStartEvent(0, "new_call", "success_tool"),
								makeInputJSONDeltaEvent(0, `{}`), makeContentBlockStopEvent(0),
							}}
						}
						return &mockAnthropicStream{err: err, events: []anthropic.MessageStreamEventUnion{
							makeTextDeltaEvent(0, text), makeContentBlockStopEvent(0),
						}}
					}
					client = c
				case "bedrock":
					c := &BedrockClient{model: "test", maxRetries: 1, contextWindowTokenCount: 1}
					c.streamImpl = func(_ context.Context, params *bedrockruntime.ConverseStreamInput) (bedrockStream, error) {
						tool, text, err := next(params)
						if err != nil {
							return nil, err
						}
						if tool {
							return &mockBedrockStream{events: []brtypes.ConverseStreamOutput{
								makeBedrockToolUseStart(0, "new_call", "success_tool"),
								makeBedrockToolUseDelta(0, `{}`), makeBedrockContentBlockStop(0),
							}}, nil
						}
						return &mockBedrockStream{events: []brtypes.ConverseStreamOutput{
							makeBedrockTextDelta(0, text), makeBedrockContentBlockStop(0),
						}}, nil
					}
					client = c
				case "genkit":
					c := &GenkitClient{g: &genkit.Genkit{}, model: "test", maxRetries: 1, contextWindowTokenCount: 1}
					c.streamImpl = func(_ context.Context, _ *genkit.Genkit, opts ...ai.GenerateOption) iter.Seq2[*ai.ModelStreamValue, error] {
						return func(yield func(*ai.ModelStreamValue, error) bool) {
							tool, text, err := next(nil)
							if err != nil {
								yield(nil, err)
								return
							}
							parts := []*ai.Part{ai.NewTextPart(text)}
							if tool {
								parts = []*ai.Part{ai.NewToolRequestPart(&ai.ToolRequest{Name: "success_tool", Ref: "new_call", Input: map[string]any{}})}
							}
							if text != "" && !yield(&ai.ModelStreamValue{Chunk: &ai.ModelResponseChunk{Content: []*ai.Part{ai.NewTextPart(text)}}}, nil) {
								return
							}
							yield(&ai.ModelStreamValue{Done: true, Response: &ai.ModelResponse{FinishReason: ai.FinishReasonStop, Message: &ai.Message{Role: ai.RoleModel, Content: parts}}}, nil)
						}
					}
					client = c
				case "responses":
					c := &OpenAIResponsesClient{model: "test", maxRetries: 1, contextWindowTokenCount: 1}
					c.responseStreamImpl = func(_ context.Context, params responses.ResponseNewParams, _ ...option.RequestOption) responseStream {
						tool, text, err := next(params)
						if err != nil {
							return &fakeResponseStream{err: err}
						}
						output := `[{"type":"message","id":"msg","role":"assistant","status":"completed","content":[{"type":"output_text","text":` + mustMarshalJSON(text) + `,"annotations":[]}]}]`
						if tool {
							output = `[{"type":"function_call","id":"fc","call_id":"new_call","name":"success_tool","arguments":"{}","status":"completed"}]`
						}
						return &fakeResponseStream{events: []responses.ResponseStreamEventUnion{
							mustResponseEvent(t, `{"type":"response.completed","sequence_number":1,"response":{"id":"resp","object":"response","output":`+output+`}}`),
						}}
					}
					client = c
				}
				registry := tools.NewRegistry()
				if err := registry.Register(&successTool{}); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				events, err := client.StreamChat(ctx, messages, registry, core.StreamOptions{DisableAutoCompaction: mode == "disabled"})
				if err != nil {
					t.Fatal(err)
				}
				var applied, failed, cancelled, done bool
				for event := range events {
					switch event.Type {
					case core.StreamEventTypeAutoCompactionApplied:
						applied = true
					case core.StreamEventTypeAutoCompactionFailed:
						failed = true
					case core.StreamEventTypeAutoCompactionCancelled:
						cancelled = true
					case core.StreamEventTypeDone:
						done = true
					case core.StreamEventTypeError, core.StreamEventTypeIncomplete:
						t.Errorf("unexpected terminal event: %+v", event)
					}
				}
				if ctx.Err() != nil {
					t.Fatal(ctx.Err())
				}
				if !done || applied != (mode == "success") || failed != (mode == "failure") || cancelled != (mode == "cancelled") {
					t.Fatalf("unexpected events: done=%v applied=%v failed=%v cancelled=%v", done, applied, failed, cancelled)
				}
				if !reflect.DeepEqual(backing, before) {
					t.Fatal("streaming mutated input history, nested tool memory, or spare capacity")
				}
				wantCalls := 3
				if mode == "disabled" {
					wantCalls = 2
				}
				if calls != wantCalls {
					t.Fatalf("requests = %d, want %d", calls, wantCalls)
				}
				if mode != "disabled" && provider != "genkit" && (!strings.Contains(compactionInput, "retained") || !strings.Contains(compactionInput, "success_tool")) {
					t.Fatal("compaction request missing historical or newly executed tool result")
				}
			})
		}
	}
}
