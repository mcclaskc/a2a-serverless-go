# a2a-serverless

This AGENTS.md file is a live document and changes as the project progresses.  Check it frequently for updates.  Also make your own updates but ask permission first.

## Project Overview
a2a-serverless is a module that allows implementation of a2a servers and clients in serverless environments. Read the readme so you have an understanding what the humans do: README.md

## Development and testing
Read and update the development and testing section in the README.md so humans can also understand it.

## Code style guidelines
- Prefer procedural over OO and Functional (hence golang).
- Functions should be pure whenever possible.
- Side effects should be isolated within single functions.
- When multiple side effects are required, they should be managed inside a bigger, 'orchestrator' function with clear function descriptions, variable names, and a clear order of operations.
- Comments should always explain 'why' not 'how' or 'what'. 
- You are a grug-brained developer: https://grugbrain.dev/
    - COMPLEXITY BAD!
    - You might say that a2a on serverless is inherently complex.  Yes, it is.  All the more reason to make it as simple and understandable as possible, and the development process should also be as simple and understandable as possible.

## Architecture
- Use the official a2a sdk: https://github.com/a2aproject/a2a-go
- Use go-service/cmd/lambda to wrap our lambda handler code
- Read our architecture decision records (ARD): TODO.md
