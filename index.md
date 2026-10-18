---
hide:
  - navigation
  - toc
---

<div class="hero-section" markdown="1">

<span class="hero-tagline">Open Source · CLI Agent</span>

# Keen Code

Your terminal-native AI coding agent — lean on context, multi-provider by design, and extended through skill-driven MCP.

<div class="hero-badges" markdown="1">

[![GitHub stars](https://img.shields.io/github/stars/mochow13/keen-code?style=social)](https://github.com/mochow13/keen-code) [![npm version](https://img.shields.io/npm/v/keen-code)](https://www.npmjs.com/package/keen-code) [![License: MIT](https://img.shields.io/github/license/mochow13/keen-code)](https://github.com/mochow13/keen-code/blob/main/LICENSE)

</div>

<div class="hero-buttons">
<a href="about.html#install-keen-code" class="md-button md-button--primary">Get Started</a>
<a href="TOUR.html" class="md-button md-button--secondary">A Tour of Keen</a>
</div>

</div>



<div class="section-divider"></div>

<div class="terminal-window">
  <div class="terminal-header">
    <span class="terminal-dot terminal-dot--red"></span>
    <span class="terminal-dot terminal-dot--yellow"></span>
    <span class="terminal-dot terminal-dot--green"></span>
    <span class="terminal-title">keen-code — ~/my-project</span>
  </div>
  <img class="terminal-body" src="assets/demo.gif" alt="Keen Code demo" />
</div>

<div class="section-divider"></div>

<div class="landing-section" markdown="1">

## Get running in a minute

Install it, point it at a project, and ask for something real.

=== "npm"

    ```bash
    npm install -g keen-code
    ```

=== "curl"

    ```bash
    curl -fsSL https://raw.githubusercontent.com/mochow13/keen-code/main/scripts/install.sh | bash
    ```

=== "try without installing"

    ```bash
    npx keen-code --version
    ```

Then start Keen in your project and give it a first task:

```bash
cd ~/my-project && keen
```

> "Explain this project and point out the riskiest file to change."

</div>

<div class="section-divider"></div>

<div class="providers-section providers-section--plain" markdown="1">

## Use any model

One tool, eleven providers — switch with `/model`, no lock-in.

<div class="providers-grid">
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/anthropic.svg" alt="Anthropic logo" /><span class="provider-name">Anthropic</span><span class="provider-models">Claude Sonnet · Opus</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/openai.svg" alt="OpenAI logo" /><span class="provider-name">OpenAI</span><span class="provider-models">GPT</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/googlegemini.svg" alt="Google Gemini logo" /><span class="provider-name">Google</span><span class="provider-models">Gemini</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/kimi.svg" alt="Moonshot AI Kimi logo" /><span class="provider-name">Moonshot AI</span><span class="provider-models">Kimi</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/deepseek.svg" alt="DeepSeek logo" /><span class="provider-name">DeepSeek</span><span class="provider-models">DeepSeek V4</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/opencode.svg" alt="OpenCode logo" /><span class="provider-name">OpenCode Go</span><span class="provider-models">Multi-model gateway</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/zai.svg" alt="Z.ai logo" /><span class="provider-name">Z AI</span><span class="provider-models">GLM</span></a>
<a class="provider-card" href="docs/ai-providers.html"><img class="provider-logo" src="assets/providers/minimax.svg" alt="MiniMax logo" /><span class="provider-name">MiniMax</span><span class="provider-models">MiniMax M</span></a>
</div>

[See all 11 providers :material-arrow-right:](docs/ai-providers.html)

</div>

<div class="section-divider"></div>

<div class="landing-section" markdown="1">

## Everything in the box

</div>

<div class="feature-grid feature-grid--bento">

<a href="docs/jev-classifier.html" class="feature-card feature-card--large">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><path d="M9 1v3M15 1v3M9 20v3M15 20v3M1 9h3M1 15h3M20 9h3M20 15h3"/></svg>
</div>
<h3>Smart Router</h3>
<p>Jev classifies every turn as simple, standard, or complex — `/model router` switches models to match, so you stop spending flagship tokens on trivial questions.</p>
</a>

<a href="docs/multi-agent-orchestration.html" class="feature-card feature-card--large">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/></svg>
</div>
<h3>Multi-Agent</h3>
<p>Delegate up to 10 bounded tasks to parallel subagents — research, implement, and review at the same time, then merge the results.</p>
</a>

<a href="docs/ai-providers.html" class="feature-card">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a4 4 0 0 1 4 4v1a1 1 0 0 0 1 1h1a4 4 0 0 1 0 8h-1a1 1 0 0 0-1 1v1a4 4 0 0 1-8 0v-1a1 1 0 0 0-1-1H6a4 4 0 0 1 0-8h1a1 1 0 0 0 1-1V6a4 4 0 0 1 4-4z"/></svg>
</div>
<h3>Multi-Provider AI</h3>
<p>Swap between Gemini, OpenAI, Anthropic, DeepSeek, and more — no lock-in.</p>
</a>

<a href="docs/mcp-skills.html" class="feature-card">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1.08-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51-1.08 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1.08 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1.08z"/></svg>
</div>
<h3>Skill-driven MCP</h3>
<p>Servers become `mcp:&lt;server&gt;` skills — prompt-efficient discovery, schemas on demand.</p>
</a>

<a href="docs/skills-system.html" class="feature-card">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
</div>
<h3>Skills System</h3>
<p>Custom slash commands that run specialized sub-agents — review, security, refactor.</p>
</a>

<a href="docs/tools.html" class="feature-card">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>
</div>
<h3>Built-in Tools</h3>
<p>Read, write, hash-anchored edits, glob, grep, bash, web fetch — plus ask and delegate.</p>
</a>

<a href="docs/turn-memory.html" class="feature-card">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/><circle cx="12" cy="12" r="4"/></svg>
</div>
<h3>Turn Memory</h3>
<p>Lean cross-turn context — compact summaries instead of raw tool traces.</p>
</a>

<a href="docs/permission-system.html" class="feature-card">
<div class="feature-card__icon">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
</div>
<h3>Permission Guards</h3>
<p>Filesystem and sensitive-path checks on every file op — the agent asks before it touches.</p>
</a>

</div>

<div class="section-divider"></div>

<div class="cta-section" markdown="1">

<p>Ready to let AI handle the boring parts?</p>
<div class="hero-buttons">
<a href="https://github.com/mochow13/keen-code" class="md-button md-button--primary">★ Star on GitHub</a>
</div>

<div class="cta-links">
<a href="TOUR.html">Tour</a>
<a href="CONTRIBUTING.html">Contribute</a>
<a href="CHANGELOG.html">Changelog</a>
</div>
</div>
