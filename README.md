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
    version: v0.1.2
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

# Конфигурация
Линтер может быть сконфигурирован для кастомных паттернов проверки потенциально чувствительных данных в логах.
Для этого укажите их в настройках golangci-lint:
```yaml
#...
settings:
  custom:
    loglinter:
      sensitivePatterns:
        - token
        - password
```

# Пример работы
<img width="1467" height="896" alt="изображение" src="https://github.com/user-attachments/assets/ab3c0cc9-92ed-48de-8470-638a0a41724f" />
Файл на котором проходила проверка расположен в ./analyzers/log-linter/testdata
