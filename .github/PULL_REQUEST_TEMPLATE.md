name: Pull Request
about: Submit code changes to IssGo
title: ''
labels: ''
assignees: ''
body:
  - type: markdown
    attributes:
      value: |
        ## 🚀 Pull Request

        Thanks for contributing! Please fill out the sections below.

  - type: dropdown
    id: type
    attributes:
      label: Type of Change
      options:
        - Bug fix
        - New feature
        - New tool
        - Documentation
        - Refactor / code quality
        - Performance improvement
        - Test
        - Other
    validations:
      required: true

  - type: textarea
    id: summary
    attributes:
      label: Summary
      description: What does this PR do? Why is it needed?
    validations:
      required: true

  - type: textarea
    id: testing
    attributes:
      label: Testing
      description: How did you test this? What commands did you run?
      placeholder: |
        ```bash
        go test ./... -v
        go build -o issgo .
        ./issgo run "test task"
        ```
    validations:
      required: true

  - type: textarea
    id: screenshots
    attributes:
      label: Screenshots / Logs
      description: If applicable, paste terminal output or screenshots

  - type: checkboxes
    id: checklist
    attributes:
      label: Checklist
      options:
        - label: Code follows project style (`gofmt -s -w .`)
          required: true
        - label: Tests pass locally (`go test ./...`)
          required: true
        - label: I have added/updated relevant tests
          required: false
        - label: Documentation updated if needed
          required: false
        - label: No new warnings from `go vet`
          required: true

  - type: input
    id: related
    attributes:
      label: Related Issues
      description: "Closes #123, Fixes #456"
