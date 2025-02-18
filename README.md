# Design Info Price Logger

This is used for scrappering data from Wishlist at designinfo.in. Using this data historical price chart is generated.

### Installation

```
export PB_APP_NAME="Price Logger"
export PB_ADMIN_NAME="Admin"
export PB_ADMIN_EMAIL="admin@example.com"
export PB_ADMIN_PASSWORD="somepassword"
export PB_APP_URL=""
export PB_DATA_ENCRYPTION_KEY=""
export OS_APP_ID="xxxxxxxxxxxxx"
export OS_APP_KEY="os_v2_app_xxxxxxxx"
export OS_TEMPLATE_ID="xxxxxxxxxxx"
export OS_SEGMENT="Total Subscriptions"

```

```
make init
make run
```

The app will be available at http://localhost:8090
