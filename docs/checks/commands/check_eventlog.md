---
title: eventlog
---

## check_eventlog

Checks the windows eventlog entries.

Basically, this check wraps this wmi query:
SELECT
	ComputerName,
	LogFile,
	Category,
	EventCode,
	EventIdentifier,
	EventType,
	Message,
	SourceName,
	Type,
	TimeWritten,
	TimeGenerated
FROM
	Win32_NTLogEvent

See https://learn.microsoft.com/en-us/previous-versions/windows/desktop/eventlogprov/win32-ntlogevent for
a description of the provided fields.


- [Examples](#examples)
- [Argument Defaults](#argument-defaults)
- [Attributes](#attributes)

## Implementation

| Windows            | Linux | FreeBSD | MacOSX |
|:------------------:|:-----:|:-------:|:------:|
| :white_check_mark: |       |         |        |

## Examples

### Default Check

    check_eventlog
    OK - Event log seems fine

### Example using NRPE and Naemon

Naemon Config

    define command{
        command_name         check_nrpe
        command_line         $USER1$/check_nrpe -H $HOSTADDRESS$ -n -c $ARG1$ -a $ARG2$
    }

    define service {
        host_name            testhost
        service_description  check_eventlog
        use                  generic-service
        check_command        check_nrpe!check_eventlog!filter=provider = 'Microsoft-Windows-Security-SPP' and id = 903 and message like 'foo'
    }

## Argument Defaults

| Argument      | Default Value                                   |
| ------------- | ----------------------------------------------- |
| filter        | level in ('warning', 'error', 'critical')       |
| warning       | level = 'warning' or problem_count > 0          |
| critical      | level in ('error', 'critical')                  |
| empty-state   | 0 (OK)                                          |
| empty-syntax  | %(status) - No entries found                    |
| top-syntax    | %(status) - %(count) message(s) %(problem_list) |
| ok-syntax     | %(status) - Event log seems fine                |
| detail-syntax | %(file) %(source) (%(message))                  |

## Check Specific Arguments

| Argument         | Description                                                            |
| ---------------- | ---------------------------------------------------------------------- |
| file             | File to read (can be specified multiple times to check multiple files) |
| log              | Alias for file                                                         |
| scan-range       | Sets time range to scan for message (default is 24h)                   |
| timezone         | Sets the timezone for time metrics (default is local time)             |
| truncate-message | Maximum length of message for each event log message text              |

## Attributes

### Filter Keywords

these can be used in filters and thresholds (along with the default attributes):

| Attribute       | Description                                         |
| --------------- | --------------------------------------------------- |
| computer        | Which computer generated the message (ComputerName) |
| file            | The logfile name                                    |
| log             | Alias for file                                      |
| id              | Eventlog id (EventCode)                             |
| eventidentifier | Event identifier (EventIdentifier)                  |
| level           | Severity level (lowercase Type)                     |
| message         | The message as a string                             |
| source          | The source system (SourceName)                      |
| provider        | Alias for source                                    |
| written         | Time of the message being written( TimeWritten)     |
