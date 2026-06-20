# third_party

Service-local внешние `.proto`-зависимости для генерации Go-кода, grpc-gateway и OpenAPI.

Сейчас здесь используются файлы `google/api`:

- `annotations.proto`
- `http.proto`
- `httpbody.proto`

Если файлы потерялись, восстановить их можно из каталога сервиса:

```bash
make proto:deps
```

Версия googleapis закреплена через `GOOGLEAPIS_REF` в service-local `Makefile`.