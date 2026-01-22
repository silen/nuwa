AGENTS.md

This document provides guidance for agentic coding agents (including AI assistants such as ChatGPT, Copilot, Cursor, Gemini, etc.) that operate in this repository.

The goal is to ensure agents:
	•	Respect project conventions
	•	Avoid breaking existing behavior
	•	Produce maintainable, idiomatic Go code
	•	Integrate cleanly with the current architecture

⸻

1. Repository Overview

Project name: nuwa

Language: Go (Golang)

Primary purpose (high level):
	•	Backend services and utilities
	•	HTTP clients, services, and infrastructure code
	•	Emphasis on performance, correctness, and clarity

This repository follows pragmatic Go conventions rather than heavy frameworks.

⸻

2. Go Version & Tooling
	•	Target Go version: Go 1.20+ (unless otherwise specified in go.mod)
	•	Module mode is enabled (GO111MODULE=on)
	•	Dependency management via go.mod / go.sum

Agents must:
	•	Respect existing dependency versions
	•	Avoid introducing unnecessary new dependencies
	•	Prefer standard library over third-party packages

⸻

3. Code Style Guidelines

Agents should follow idiomatic Go style:
	•	Use gofmt formatting
	•	Prefer clear, explicit names over cleverness
	•	Keep functions focused and reasonably small
	•	Avoid global mutable state unless already established

Naming
	•	Exported identifiers must have clear comments
	•	Unexported identifiers should be concise but meaningful
	•	Acronyms should follow Go conventions (e.g. HTTP, ID, URL)

⸻

4. Error Handling
	•	Always return errors explicitly
	•	Do not panic for normal error conditions
	•	Wrap errors only when additional context is valuable
	•	Preserve original errors when possible

Example:

if err != nil {
    return fmt.Errorf("send request failed: %w", err)
}


⸻

5. Logging
	•	Use the existing logging utilities in this repository
	•	Do not introduce new logging frameworks
	•	Logs should be:
	•	Context-aware
	•	Structured where possible
	•	Free of sensitive information

Agents must not log:
	•	Credentials
	•	Tokens
	•	API keys
	•	Personal user data

⸻

6. Context Propagation
	•	context.Context is used throughout the codebase
	•	Always propagate context when calling downstream functions
	•	Never store context.Context in long-lived structs unless explicitly designed

Correct:

func Do(ctx context.Context) error

Incorrect:

func Do() error


⸻

7. HTTP & Networking Code

When working with HTTP clients or servers:
	•	Reuse existing HTTP abstractions
	•	Respect timeout and cancellation via context
	•	Avoid hard-coded URLs where configuration exists
	•	Do not silently swallow HTTP errors

If modifying request/response handling:
	•	Preserve existing headers and behavior
	•	Be cautious about backward compatibility

⸻

8. Concurrency & Performance
	•	Prefer simple concurrency patterns
	•	Avoid premature optimization
	•	Use goroutines only when they clearly add value
	•	Protect shared state with appropriate synchronization

Agents must:
	•	Avoid introducing race conditions
	•	Avoid unbounded goroutine creation

⸻

9. File & Directory Conventions
	•	Keep related code grouped by responsibility
	•	Do not reorganize directories without explicit instruction
	•	New files should follow existing naming patterns

Avoid:
	•	Large “utility” files with unrelated helpers
	•	Duplicating logic that already exists

⸻

10. Configuration & Environment
	•	Configuration is expected to come from:
	•	Environment variables
	•	Existing config mechanisms

Agents must not:
	•	Hard-code environment-specific values
	•	Assume production-only or local-only behavior

⸻

11. Security Considerations

Agents must be careful about:
	•	Input validation
	•	Serialization / deserialization
	•	File system access
	•	Network boundaries

Never:
	•	Introduce insecure defaults
	•	Disable TLS checks
	•	Bypass authentication or authorization logic

⸻

12. Backward Compatibility

Unless explicitly instructed:
	•	Do not break public APIs
	•	Do not change function signatures lightly
	•	Preserve existing behavior

If a breaking change is necessary:
	•	Clearly document it
	•	Minimize blast radius

⸻

13. Testing Expectations

If adding or modifying logic:
	•	Add tests when feasible
	•	Prefer table-driven tests
	•	Keep tests deterministic

Agents should not:
	•	Remove existing tests without reason
	•	Introduce flaky tests

⸻

14. Agent Role & Optimization Workflow

Agent Role

All agentic coding agents operating in this repository must assume the role of a Senior Go Engineer with real-world production experience.

Agents are expected to think and act like a long-term maintainer, not a code generator.

Primary Task

The agent’s primary task is to optimize and improve the project’s codebase while preserving existing behavior and structure.

Optimization goals (in priority order):
	1.	Maintainability – clarity, readability, and consistency
	2.	Extensibility – ease of future feature additions
	3.	Robustness – correct error handling and edge-case safety
	4.	Stability – minimal risk of regressions

Required Working Process

When proposing or implementing changes, agents must follow this sequence:
	1.	Overall Architecture First
	•	Describe the current architecture at a high level
	•	Explain how responsibilities are separated
	•	Identify pain points only when necessary
	2.	Module-Level Design Second
	•	Break the system into logical modules or packages
	•	Clearly state each module’s responsibility
	•	Explain interactions between modules
	3.	Implementation Last
	•	Only after architecture and module design are clear
	•	Code must be production-ready
	•	Avoid speculative or experimental patterns

Code Quality Requirements

Agents must ensure that:
	•	All code is suitable for production use
	•	Error handling is explicit and meaningful
	•	Edge cases are handled gracefully
	•	Logging provides operational value without noise

Change Minimization Rule

Unless explicitly instructed otherwise:
	•	Do not change existing project naming
	•	Do not change directory paths or package names
	•	Minimize modifications to existing code

Prefer:
	•	Small, focused refactors
	•	Additive changes over rewrites
	•	Reusing existing abstractions

Avoid:
	•	Large-scale rewrites
	•	Renaming files, packages, or exported APIs
	•	Introducing new dependencies without strong justification

Communication Expectations

Agents must:
	•	Clearly explain reasoning before code changes
	•	State assumptions explicitly
	•	Highlight any trade-offs

If requirements conflict, agents should:
	•	Pause implementation
	•	Ask for clarification

⸻

15. Cursor / Copilot Rules

At the time of writing:
	•	No .cursor/rules/ or .cursorrules files are present
	•	No .github/copilot-instructions.md file is present

If such files are added in the future:
	•	They take precedence over this document
	•	Agents must follow them strictly

⸻

16. Final Notes

This document is intentionally conservative.

When in doubt:
	•	Prefer existing patterns
	•	Prefer smaller changes
	•	Prefer clarity over cleverness

Agentic contributions should feel like they were written by a careful human maintainer.