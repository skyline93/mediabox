# mediabox

## Getting started

```bash
sudo apt-get update
sudo apt install libvips libvips-dev
export CGO_ENABLED='1'
```

```bash
cd frontend
npm install
npm run serve
```

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

```bash
make
```

```json
cat <<EOF > config.json
{
    "http_host": "127.0.0.1",
    "http_port": 8000,
    "use_tls": true,
    "tls_cert": "server.crt",
    "tls_key": "server.key",
    "db_driver": "postgresql",
    "db_dsn": "host=127.0.0.1 user=mediabox password=123456 dbname=mediabox port=5432 sslmode=disable TimeZone=Asia/Shanghai",
    "storage_path": "storage"
}
EOF
```

```bash
./mediabox start --config config.json
```

## frontend

```bash
vue create frontend
```
