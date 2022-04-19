## Порядок настройки Prometheus и Grafana
Для запуска Prometheus  и Grafana было принято решение использовать docker  
Листинг
```
version: '3.8'

volumes:
    prometheus-data:
        driver: local
    grafana-data:
        driver: local

services:
    prometheus:
        image: prom/prometheus:latest
        container_name: prometheus
        ports:
            - "9090:9090"
        volumes:
            - /etc/prometheus:/etc/prometheus
            - prometheus-data:/prometheus
        restart: unless-stopped
        command:
            - "--config.file=/etc/prometheus/prometheus.yml"
    grafana:
        image: grafana/grafana-oss:latest
        container_name: grafana
        ports:
            - "3000:3000"
        volumes:
            - grafana-data:/var/lib/grafana
        restart: unless-stopped
```
В первой строке листинга определяется версия docker-compose файла. Затем определяются сервисы, для данного приложения будут использоваться два сервиса prometheus и grafana. Следующий тег image используется, чтобы определить исходный образ, были использованы подготовленные официальные образы prometheus и grafana. Далее тегом environment определяются переменные окружения. С помощью тега ports пробрасываются порты.
## Порядок запуска
Для запуска необходим переместиться в директорию, в которой находится docker-compose файл и запустить команду docker-compose up. Для подключения к grafana через любой браузер необходимо перейти по адресу http://loclhost:3000, для входа необходимо использовать admin логин и пароль admin. Далее необходимо создать подключение к серверу БД. В качестве имени узла используется имя контейнера, так как оно является именем узла в сети docker-compose.
