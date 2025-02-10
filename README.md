# Design Info Price Logger

This is used for scrappering data from Wishlist at designinfo.in. Using this data historical price chart is generated.

### Dependencies

- ClickHouse DB Server
  - .env file variables: DB_HOST, DB_PORT, DATABASE, DB_USERNAME, DB_PASSWORD
- goDotEnv module

### Installation

```
make build
cd dist
sudo ./install
```
