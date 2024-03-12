# Quest

[![Build status](https://img.shields.io/github/actions/workflow/status/Fabian-G/quest/test.yml)](https://github.com/Fabian-G/quest/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Fabian-G/quest)](https://goreportcard.com/report/github.com/Fabian-G/quest)
[![Release](https://img.shields.io/github/v/release/Fabian-G/quest)](https://github.com/Fabian-G/quest/releases)
[![GitHub](https://img.shields.io/github/license/Fabian-G/quest)](https://github.com/Fabian-G/quest/blob/main/LICENSE)

Quest is a command line interface for [todo.txt](https://github.com/todotxt/todo.txt) with many additional features:

- [Views](https://fabian-g.github.io/quest/views)
- Powerful [query language](https://fabian-g.github.io/quest/selection)
- Time tracking with [Timewarrior](https://github.com/GothenburgBitFactory/timewarrior)
- [Recurrence](https://fabian-g.github.io/quest/recurrence)
- Due/Threshold dates (as a byproduct of the views feature)
- Multi line [notes](https://fabian-g.github.io/quest/notes)
- [Tag Expansions](https://fabian-g.github.io/quest/tag-expansions) (e.g. expand `tomorrow` to the date of tomorrow)
- Sublist [editing](https://fabian-g.github.io/quest/editing) with your favourite editor

## Installation

To build from source:
```bash
go install 'github.com/Fabian-G/quest@latest'
```

or download the precompiled binary from the [Release Page](https://github.com/Fabian-G/quest/releases).

## Why Quest?

The purpose of this README is to show an example usage of quest.
To learn how things work checkout the [documentation](https://fabian-g.github.io/quest).
Many things that are shown in the following examples are not possible with the
default configuration. If you want the exact behaviour shown here choose *Dev's Choice* 
when running `quest init`.

### Clearing the Inbox

A good way to start the day is by clearing out the task inbox, which contains
tasks that need refinement (like rewriting it to be actionable, assigning due dates, projects or priorities).
First let's see what currently is in the inbox.
```bash
~ ❯ quest inbox     
 #  Created On  Description 
 1  2024-03-11  Plan Bilbos birthday 
 2  2024-03-11  Take the hobbits to Isengart
 3  2024-03-11  Get stabbed by Morgul-knife      
 4  2024-03-11  Meet Gandalf      
 5  2024-03-11  Simply walk into Mordor 
```

Often times it is most efficient to edit all the inbox tasks at once in your editor:
```bash
~ ❯ quest inbox edit
```

This will open the 5 inbox items in your favourite editor in todo.txt format. 

With this configuration a task item is considered an inbox item if it does not have a priority assigned.
So after assigning at least a priority to each item the inbox should be cleared:
```bash
~ ❯ quest inbox     
no matches
```

### Planning the day

The *next* view will show us all actionable tasks sorted by what Quest thinks should be done next
(based on priority and due dates).
```bash
~ ❯ quest next     
 #  Score  Priority  due        Projects  Contexts  Description 
 5  7.4    (D)       in 1 day   +ring               Simply walk into Mordor       
 4  7.1    (A)                                      Meet Gandalf            
 2  6.6    (D)       in 7 days                      Take the hobbits to Isengart      
 3  5.5    (B)                            @errands  Get stabbed by Morgul-knife            
 1  1.0    (E)                  +bday               Plan Bilbos birthday
```

Personally I like to pick a few tasks from this view and schedule them for the day.
```bash
~ ❯ quest set -a do:today on 5,1,3
Set tag "do" to "today" on 3 items
```
Note that the word "today" will actually be expanded to the date of today.

Now they will show up in the *today* view.
```bash
~ ❯ quest today
 #  Score  Priority  due        Projects  Contexts  Description 
 5  7.4    (D)       in 1 day   +ring               Simply walk into Mordor       
 3  5.5    (B)                            @errands  Get stabbed by Morgul-knife            
 1  1.0    (E)                  +bday               Plan Bilbos birthday
```

Other tasks may need to be delegated, because we can't do them ourselves.
```bash
~ ❯ quest set @waiting on 2
Set context "@waiting" on item #2: Take the hobbits to Isengart due:2024-03-18 @waiting
```

### Start working on a task

Since Task #5 is due tomorrow we should better start working on it.
```bash
~ ❯ quest track 5
Started tracking item #5: Simply walk into Mordor +ring due:2024-03-12 do:2024-03-11 tr:28502833
```
By running `timew` you should see that tracking has been started.
```bash
~ ❯ timew
Tracking "Simply walk into Mordor" ring
  Started 2024-03-11T16:14:10
  Current               15:09
  Total               0:00:59
```

### Adding notes to a task

Word has it that one does not simply walk into Mordor. 
So maybe we have some difficulties with task 5 and want to take some notes on that.
With the following command Quest will create a markdown file, attach it to 
task 5 and then open it for editing.
```bash
~ ❯ quest notes 5
```

### Quick capture 

When tasks pop up while we actually want to be focused on the current task
we can easily get them out of our head by using the `add` command.
```bash
~ ❯ quest add Eat second breakfast
Added task #6
```
Task #6 will automatically show up in the inbox view, because it does not have a priority.

### Completing tasks

When we are done with task #5 we can just use `complete` (e.g. with a word from the target task).
```bash
~ ❯ quest complete mordor
Completed item #5: Simply walk into Mordor +ring due:2024-03-12 tr:28502833 n:y362
```
It is safe to be rather vague about which task to complete, because if there are multiple 
matches Quest will ask you which one you mean.

### Postponing the rest

Plans don't always work out as expected. If there are tasks left undone at the end of the day we can
easily schedule them for tomorrow.
```bash
~ ❯ quest today set -a do:tomorrow 
Set tag "do" to "tomorrow" on 2 items
```

