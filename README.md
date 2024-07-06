# Coffeecup CLI

## Commands

### login

`cc login`

Asks for company URL, username, and password. Stores credentials somewhere safe to re-login if necessary.

### start

`cc start`

Starts a new time entry (or resumes the last active one if not specified). If starting a new time entry for a project that already has a time entry of the same day, the existing time entry will be resumed.

#### Options

- project (default: last used, shows list of available projects on demand)
- task (default: previously used if available for project, shows list of available tasks on demand)

#### Examples

`cc start taly`

### stop

`cc stop`

Stops the currently running time entry. If there is none, doesn't do anything (maybe print warning)

### status

`cc status`

Shows if there is a currently running time entry and the time entries of today

| Project | Duration | Comment                   |
| :------ | :------- | :------------------------ |
| taly    | 2:13     | - reviews                 |
| ctr     | 4:30     | - reviews<br>- refinement |

### amend

`cc amend`

#### Options

- time entry (default: last)
- new duration (in minutes, prefilled with current value)
- new comment (text area, append to previous one)

### projects

`cc projects`

| ID  | Name              |
| :-- | :---------------- |
| 123 | TALY 2024 (Q3+Q4) |
| 345 | BMP Claims 2024   |

#### Sub commands

`cc projects alias {id} {alias}`

Alias can then be used in `cc start {project alias}
