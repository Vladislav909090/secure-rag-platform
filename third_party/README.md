# third_party

Внешние `.proto`-зависимости, которые нужны для генерации Go-кода и grpc-gateway.

Сейчас здесь используются файлы `google/api`:

- `annotations.proto`
- `http.proto`
- `httpbody.proto`

Если каталог пустой или файлы потерялись, восстановить их можно командой из корня репозитория:

```bash
make proto:deps
```
