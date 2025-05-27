## A FEW SMALL TIPS

1. Hide field depends on condition

I have a struct that performed a specified field
```
RefreshTokenEncrypted string     `json:"refresh_token_encrypted"`
```
however I want to show this field for a little levels or scopes. To do this, you just set a condition to this mechenism implementing
```
RefreshTokenEncrypted *string     `json:"refresh_token_encrypted,omitempty"`
```
Therefore, the final result is included the above field

2. Build docker compose
```
docker compose -f docker-compose.grafana.yml up -d
```

```
docker compose -f docker-compose.prometheus.yml up -d
```
