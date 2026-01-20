---
description: Review code, refactor based on feedback, and generate PR template
---

You are orchestrating a three-step workflow: code review → refactor → PR generation.

## Step 1: Code Review

Launch the code-reviewer agent to review the current codebase changes:

```
Use Task tool with:
- subagent_type: "code-reviewer"
- description: "Review recent code changes"
- prompt: "Review the recent code changes in this codebase against best practices, coding standards, and the original requirements. Identify any issues, bugs, or improvements needed."
```

Wait for the code-reviewer agent to complete and read the feedback.

## Step 2: Code Refactor

If the code-reviewer found issues, launch the code-refactorer agent:

```
Use Task tool with:
- subagent_type: "code-refactorer"
- description: "Implement code review feedback"
- prompt: "Implement the following feedback from the code review: [PASTE REVIEW FEEDBACK HERE]. Make the necessary changes to fix the identified issues."
```

If the code-reviewer approved the code with no issues, skip this step.

## Step 3: PR Template

Launch the pr-template-generator agent:

```
Use Task tool with:
- subagent_type: "pr-template-generator"
- description: "Create pull request template"
- prompt: "Create a pull request template for the recent changes. Include a comprehensive summary of what was changed."
```

## Summary

After all agents complete, provide:
- Code review results
- Refactoring performed (if any)
- PR template location
- Next steps (commit, push, create PR)
