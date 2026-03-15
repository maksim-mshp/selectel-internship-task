# Линтер для проверки лог-записей

Решение тестового задания на стажировку в Selectel.

**Автор:** Чернышков Максим

## Установка

Для установки линтера необходимо выполнить команду в терминале в корне проекта.

**Linux / macOS**
```sh
curl -sSfL https://raw.githubusercontent.com/maksim-mshp/selectel-internship-task/main/scripts/setup.sh | sh
```

**Windows PowerShell**
```powershell
iwr -useb https://raw.githubusercontent.com/maksim-mshp/selectel-internship-task/main/scripts/setup.ps1 | iex
```

## Использование

Для использования линтера необходимо собрать кастомный `golangci-lint`.

```sh
golangci-lint custom -v
```

Linux / macOS: `./custom-gcl run` \
Windows: `.\custom-gcl.exe run`

### Автоисправление

Можно запустить `custom-gcl run --fix` чтобы применить все исправления сразу. Также можно увидеть предлагаемые исправления в JSON `custom-gcl run --output.json.path=stdout` (они будут закодированы в base64).

## Тестирование

```bash
go test -v ./...
```

## Пример использования

![](https://i.imgur.com/4cwEoIZ.png)