# Autentikasi Fiber

This is the authentication microservice (Fiber + Go).

## Run for local network (LAN) access

By default the server binds to the address set in `APP_HOST` and port `APP_PORT`.
To allow other devices on your local network to reach the service, set:

```
APP_HOST=0.0.0.0
APP_PORT=3000
```

You can place those vars in a `.env` file (we use `github.com/joho/godotenv` in `main.go`). There's an example `.env.example` included.

After that start the server:

```powershell
# from the autentikasi-fiber folder
go run .
```

From another machine on the same LAN, use the host machine's local IP (for example `192.168.1.42`) and the configured port:

http://192.168.1.42:3000

Security note: Binding to 0.0.0.0 exposes the service on all interfaces. Only do this on trusted networks or behind a proper firewall. Use authentication/HTTPS in production environments.

## Dev seeding (optional)

For development convenience there's an optional seeder that will create a developer user and a few sample transactions.

To enable it, set the environment variable before starting the service:

```powershell
# enable dev seeding (creates dev@example.com with password "password")
$env:DEV_SEED = 'true'
go run .
```

The seeder is safe: it won't duplicate the dev user or overwrite existing transactions for that user.
