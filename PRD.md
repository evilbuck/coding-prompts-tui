# LLM Coding Prompt Builder

## Purpose

A tool to help a developer build a prompt with code context.
Take an existing projects files, and selected files, and creating
purposeful prompts to give to an agent or a chatgpt, gemini, or claude type
chat program for better results that are directly related to your codebase.

**Tooling**
A command line tool built with go and bubble tea.

## MVP Focus

Create a tui interface that will utilize the cwd as a basis for selecting the 
files for context.

### Use Case

As a user, I open the code-prompt with `cd my-project-dir`, `prompter .`
I want to generate a prompt to paste into claude desktop. This prompt 
should have context about the files and directories in my code base.

I want claude to have a tree view of the files available. I also want 
claude to have the full file contents of files I select from the tui interface.

The prompt will be built with an easy to understand xml structure that identifies
the file tree structure of the project, the file contents (of the selected files) 
and which files they are from the tree. It should also have a system prompt and user prompt available. The system prompt will be built-in and configurable. For now just use the default from the personas folder, `personas/default.md`.

The user prompt will be taken from the user in the tui interface.

## TUI

**MVP**
Different tiled areas. Using [tab] will select the different areas.

**Chat Area**
This area is where the user prompt will come from. The user will be able to tab into this area and write their prompt for the LLM.

**File Selection**
This is a tree list of the files and folders.
The user will be able to navigate with the up, down arrows.
Folders will have a folder icon in front of the name. Files a file name. folders will have a checkbox in front of the name. 
Pressing spacebar will select the file or the folder and it's contents.

**Selection**
Selection will add the files to the prompt context. Each file will have an xml node notating which file it is from the root of the project path.
Do not allow checking folders. We want the file selection to be explicit for MVP.

## XML structure

This is a proposal, and I'm open to suggestions with reason. The purpose is to make it clear as day for the LLM to process and understand the limited scope of the project.

```xml
<filetree></filetree>
<file name="app/controllers/tasks_controller.rb">...filecontents</file>
<SystemPrompt>
You are a seasoned engineer. You are an expert using ruby on rails. 
You keep tasks simple and succinct. You use canonical solutions that are easy to understand.
You prefer to build only what is necessary and iterating for more complex solutions.
</SystemPrompt>
<UserPrompt>
Help me design a todo list given the files. I need to know how to structure it.
  Can we make this without reloading the page when I check a task item as completed?
</UserPrompt>
```
