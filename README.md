# Timeout Checker

This is a helper program for the azurerm Terraform provider. It checks whether the timeouts in the defintions match what is in the documentation.

## Usage

```shell
$ go install github.com/wyattfry/timeout_checker
$ timeout_checker ~/terraform-provider-azurerm/internal/services/mysql/mysql_flexible_database_resource.go ~/terraform-provider-azurerm/
Checking Timeouts...
Found matching documentation file...
+--------+--------+--------+
|        |  DEF   |  DOCS  |
+--------+--------+--------+
| Create | 1h0m0s | 1h0m0s |
| Read   | 5m0s   | 5m0s   |
| Update | 0s     | 0s     |
| Delete | 1h0m0s | 1h0m0s |
+--------+--------+--------+
* Documentation Matches! *
```