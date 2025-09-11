# Saving Context and Chats

## feature

I want a way to save the state of a "session". The state includes the chat prompt, the selected files, and the project directory.

I want a menu item to be able to save this state.
When initialized, a minimal floating panel with a text input labeled "name" should appear. Pressing enter should save the chat with the name. The name field should not be blank.

**Interface**
When activated, this should be a popover dialog with a list of chats.
The chat list can be navigated with a keyboard or mouse.
Selecting a chat by name will load the chat state into the current session.
The current session is saved as "scratch" if it does not already have a name.

"q" will close the dialog.

On the bottom of the dialog, I want a footer similar to our menu. 
We should create an abstraction that we can use anywhere we want a footer menu.
Create this new component.
