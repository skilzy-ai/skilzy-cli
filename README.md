# Skilzy CLI

A command-line tool for creating, validating, and publishing AI skills to the Skilzy Registry.

## Installation

### Pre-built binaries

Download the latest release for your platform from the [releases page](https://github.com/skilzy-ai/skilzy-cli/releases).

### From source

```bash
go install github.com/skilzy-ai/skilzy-cli@latest
```

## Quick Start

```bash
# Create a new skill
skilzy init my-awesome-skill

# Validate your skill
cd my-awesome-skill
skilzy validate

# Package for distribution
skilzy package

# Login and publish
skilzy login
skilzy publish dist/my-awesome-skill-0.1.0.skill
```

## Commands

- `skilzy init <skill-name>` - Create a new skill
- `skilzy validate` - Validate skill.json and structure
- `skilzy package` - Package skill into .skill file
- `skilzy convert <path>` - Convert existing skill to Skilzy format
- `skilzy search <query>` - Search the Skilzy registry
- `skilzy login` - Authenticate with your API key
- `skilzy publish <package>` - Publish to registry
- `skilzy me whoami` - Validate your API key
- `skilzy me skills` - List your published skills

## Documentation

For full documentation, visit [docs.skilzy.ai](https://docs.skilzy.ai)

## License

MIT
