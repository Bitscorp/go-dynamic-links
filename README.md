# Create new links

```
curl -X POST http://localhost:8080/projects/myapp/links/new \
-H "Content-Type: application/json" \
-d '{"ios":"https://ios.example.com", "android":"https://android.example.com", "web":"https://web.example.com"}'
```

# Test health check

```
curl http://localhost:8080/health
```

# Test redirect (will depend on your User-Agent)

```
curl -L http://localhost:8080/myapp
```
