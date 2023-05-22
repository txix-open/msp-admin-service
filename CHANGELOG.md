### v5.0.0
* изменён механизм ролей и пользователей
### v4.3.0
* реализован метод получения списка доступных ролей
* реализован метод блокировки пользователей
* реализован метод отзыва сессии
* реализован метод получения списка сессий
* реализован сбор аудита
* реализован метод просмотра журнала аудита
* роли полученные от СУДИР сохраняются автоматически
* профиль администратора полученный от СУДИР обновляются автоматически
* реализована блокировка УЗ при неактивности больше n дней
* обновлены зависимости
* добавлены SQL метрики
### v4.2.0
* добавлена защита от брутфорса админ пароля
### v4.1.1
* испавлено помедение метода /get_profile в новом флоу
### v4.1.0
* добавлен метод аутентификации токена
* удалены бессрочные токены. Время жизни токена по умолчанию 1 час
* обновлен метод logout: закрывает все текущие сессии администратора (на всех устройствах)
### v4.0.2
* исправлена ошибка метадаты grpc 
### v4.0.1
* изменен default_remote_config
### v4.0.0
* миграция сервиса на isp-kit
* миграция на версию GO 1.19.1
* обновлен формат логов в json
* обновлена документация swagger: разделены ручки создания и обновления пользователя
### v3.5.0
* add role to response /user/get_profile
* add read_only_admin role
* add role link for sudir user
* fix sudir user response
### v3.4.1
* updated dependencies
* migrated to common local config
### v3.4.0
* add method /auth/login_with_sudir
### v3.3.1
* updated dependencies
### v3.3.0
* add `user/get_design` method
### v3.2.2
* updated isp-lib
### v3.2.1
* updated isp-lib
* updated isp-event-lib
### v3.2.0
* Added config edit sessions
* Added real-time feature with websockets
### v3.1.0
* update to go mod
### v3.0.0
* update to new isp-lib & config service
### v2.1.2
* add default remote configuration
