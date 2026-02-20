# Линтер для анализа логирования
# Описание
Линтер для анализа логирования в Go. Он проверяет сообщения в логах по следующим правилам:
- Сообщения должны начинаться со строчной буквы
- Сообщения должны использовать только латинский алфавит
- Сообщения не должны содержать спецсимволы
- Сообщения не должны содержать потенциально секретные данные

# Установка
Линтер реализован как плагин для golangci-lint. Для установки в свой проект выполните следующие шаги:
1. Создайте файл `.custom-gcl.yml`:
```yaml
version: v2.10.1
plugins:
  - module: "github.com/prr133f/go-log-linter"
    import: "github.com/prr133f/go-log-linter/plugin"
    version: v0.1.1
```
2. Включите линтер в конфигурации golangci-lint:
```yaml
# .golangci.yml
version: "2"
linters:
  default: none
  enable:
    - loglinter
  settings:
    custom:
      loglinter:
        type: "module"
        description: "Custom log linting rules"
        original-url: github.com/prr133f/go-log-linter
```
3. Соберите кастомный линтер и запустите его:
```sh
golangci-lint custom -c .custom-gcl.yml -v
./custom-gcl run -c .golangci.yml ./... 
```
