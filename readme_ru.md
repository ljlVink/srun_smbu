# srun_smbu

Большое спасибо [за эту программу с открытым исходным кодом.](https://github.com/vouv/srun)

Инструмент сетевого подключения МГУ-ППИ, вы можете указать IP-адрес сетевой карты при настройке в исходном проекте.

### Использование

```
./srun_smbu \
-username your_username \
-pass your_password \
-addr your_auth_url \
-nwip your_network_card_ip
```

username (обязательно)имя пользователя

pass (обязательно)пароль

addr (необязательно)адрес аутентификации (обычно http://172.20.5.18/)

nwip (обязательно)ip-адрес сетевой карты
