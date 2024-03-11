# Getting Started

To start using Quest you don't necessarily need any configuration. That is, if all you need
is a simple CLI to interface with your todo.txt without any special features.

In most cases you will want a config though. Since there is probably a bit of learning curve
for this, Quest comes with a few presets.
Therefore, after installing Quest the first thing you will want to do is to choose your preset by running:
```bash
quest init
```
If you don't follow a particular methodology (like GTD) it is recommended to choose 
*extended todo.txt*. This preset comes with all advanced features enabled (like recurrent tasks), but
other than that it is pretty minimal.

## Adding/Listing/Completing/Editing tasks

Adding a new task can be done by using the `add` command.
Every remaining argument will be considered part of the task description.
```bash
quest add Do the dishes
```

To list tasks, simply run `quest` without any arguments.
You can also add keywords to search for, `quest -- dish` will only list
tasks containing the substring "dish".

To complete a task you can use the `complete` command, which takes a list 
of selectors (e.g. words to search for) as well.
If you know you want to complete a task that contains the word "dish" simply
run this.
```bash 
quest complete dish
```
If multiple tasks contain the word dish Quest will ask you which tasks
you mean.

To edit a task in your editor you'll want to use the `edit` command.
```bash
quest edit dish
```

## Updating Projects/Contexts/Tags

Words in a task description that are prefixed by an @ symbol are considered contexts.
If a word is prefixed by a + Symbol it is considered a project.
The `set` and `unset` commands are used to manage these types of metadata.
For example to add the task on line 3 to the project foo, one would run:
```bash
quest set +foo on 3
```
and to remove it from the project:
```bash
quest unset +foo on 3
```

Tags are key-value pairs that carry metadata for the task.
Commonly this is used to carry due dates:
```bash
quest set due:2024-12-12 on 3
```
or to make a task recurrent:
```bash
quest set t:2024-12-10 rec:1y on 3
```

## Where to go from here

With these few commands you can probably get started already. 
At some point you will want to look into how to define [views](views.md)
for which you will need at least a basic understanding of the [query language](selection.md).
For convenience you might also want to look into [tag expansions](tag-expansions.md).
