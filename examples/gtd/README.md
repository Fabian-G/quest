# Basic GTD Configuration

## Views

- Inbox: Items that need to be processed (indicated by the `+inbox` project)
- Next: Items that are immediately actionable
- Scheduled: Items that are scheduled for a later date (indicated by the `t` tag)
- Waiting: Items that require action by someone else (indicated by the `delegated` tag)
- Someday: Items that may or may not be actionable at some point in the future (indicated by the `sm` tag)

## Adding to the inbox

All views except next automatically set the `+inbox` project.

```bash
quest add This gets added to the inbox
quest inbox add This will also be added to the inbox
```

## Declaring dependencies between tasks

The `id` and `after` tags can be used to indicate a dependency. 
```
a task with id:1
a blocked task after:1
```
"a blocked task" won't show up in the next view until the task with id 1 is completed.

## Partially ordered projects

With the `order` tag one can express that tasks should be done sequentially.
```
T1 order:1
T2 order:2
T3
```
`quest next` will return T1 and T3. After T1 is completed T2 will show up.