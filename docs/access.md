Access level
============

Goop uses a hierarchical access model. Access levels determine what commands can be accessed, whether or not messages will be relayed to other realms, and whether or not the user will be banned upon joining the channel. 

|Role        | Level|Description|
|------------|-----:|-----------|
|`owner`     | 1000 | Bot owner, has access to everything. |
|`admin`     | 300  | Administrator, has access to everything except settings. |
|`operator`  | 200  | Channel operator, can kick/ban. |
|`whitelist` | 100  | Trusted user, only kickable/bannable by admins. |
|`voice`     | 1    | Chat will be relayed between gateways. |
|`ignore`    | -1   | Ignore user, do not relay chat and do not process commands. |
|`kick`      | -100 | Auto kick. |
|`ban`       | -200 | Auto ban. |
|`blacklist` | -300 | Auto ban, only unbannable by admins. |

An access level can be assigned to a particular user (with the [.set](commands_builtin.md#set) command) or to a particular group (such as users with a certain role on Discord or users from a certain clan on Battle.net).