# Sync Folders app

## Описание

Программа позволяет выполнять полную синхронизацию между оригинальной папкой и её дубликатом.

## Установка

Для установки приложения необходимо выполнить команду

```bash
  make build
```
## Запуск

Для запуска приложения необходимо выполнить команду

```bash
  ./syncfolder <относительный путь к оригинальной папке> <относительный путь к дублирующей папке> 
```
Пример:
```bash
  ./syncfolder ./origin_folder ./duplicate 
```
    