# `request-proxy`
# MITM Proxy with TLS Interception

A lightweight Go-based MITM (Man-in-the-Middle) proxy that supports intercepting and modifying HTTP/HTTPS traffic. It dynamically generates TLS certificates based on SNI for HTTPS CONNECT tunnels, allowing full visibility and manipulation of encrypted traffic.

---

## 🔍 Features

- ✅ Intercept and modify **HTTP** requests and responses
- ✅ Handle **HTTPS CONNECT** tunneling with dynamic certificate generation
- ✅ Support for **SNI-based certificate caching**
- ✅ Modify request/response headers and body content
- ✅ Transparent proxy support using custom CA
- ✅ Lightweight and extensible architecture

---

## 🧰 Technologies Used

- Go 1.20+
- `net/http`, `crypto/tls`, `crypto/x509`, `bufio`, `io`
- Self-signed root CA for MITM interception
- Dynamic certificate generation per domain

---

## 📁 Project Structure

```
.
├── main.go                     # Entry point
├── proxy/
│   ├── proxyHandler.go         # Connection handling and protocol detection
│   ├── httpHandler.go          # HTTP(S) request processing and MITM logic
│   └── util/
│       └── certificateUtil.go  # Root & leaf certificate generation
└── model/
    └── wrapRequest.go          # Wrapper for connection and reader/writer
```


---

## 🛠 Key Components

### 1. **Certificate Management**

- Generates a self-signed **root CA certificate**
- Dynamically creates **leaf certificates** for each requested domain
- Supports **SNI-based certificate caching**
- Stores certificates in PEM format ([cert.crt](file:///Users/wushaojie/Documents/project/golang/request-proxy/cert.crt), [cert.key](file:///Users/wushaojie/Documents/project/golang/request-proxy/cert.key))

### 2. **MITM Proxy Logic**

- Detects CONNECT requests and establishes HTTPS tunnels
- Uses generated certificates to perform TLS handshake with client
- Forwards decrypted traffic to target server
- Allows modification of headers and body content

### 3. **Transparent Certificate Handling**

- Signs leaf certificates using the internal root CA
- Responds to browser trust prompts by providing trusted certificates
- Logs all certificate generation and usage

---

## ⚙️ Configuration

You can configure:

- Root certificate validity period
- Certificate cache size and TTL
- Allowed TLS versions and cipher suites
- Request/response modification hooks

All configurations are handled via code-level settings and can be extended.

---

## 🧪 Usage

### 1. Start the Proxy Server

```bash
go run main.go
```


By default, the proxy listens on `localhost:8080`.

### 2. Configure Browser or System Proxy Settings

Set your system or browser proxy to:

```
Host: 127.0.0.1
Port: 8080
```


### 3. Trust the Root Certificate

Install the generated root certificate ([cert.crt](file:///Users/wushaojie/Documents/project/golang/request-proxy/cert.crt)) into your OS/browser trust store:

#### macOS:
- Open "Keychain Access"
- Drag and drop [cert.crt](file:///Users/wushaojie/Documents/project/golang/request-proxy/cert.crt) into "System" keychain
- Trust the certificate for SSL/TLS

#### Windows:
- Run `certmgr.msc`
- Import [cert.crt](file:///Users/wushaojie/Documents/project/golang/request-proxy/cert.crt) into "Trusted Root Certification Authorities"

#### Linux (Chrome/Firefox):
- Manually import [cert.crt](file:///Users/wushaojie/Documents/project/golang/request-proxy/cert.crt) in browser settings

---

## 📈 Example Use Cases

| Use Case | Description |
|----------|-------------|
| Web Scraping | Intercept and log HTTPS requests/responses |
| API Testing | Modify request bodies, inject test payloads |
| Traffic Analysis | Log and inspect encrypted traffic |
| Debugging Tools | Inject scripts or modify HTML/CSS/JS |
| Security Research | Analyze third-party network behavior |

---

## 🧪 Sample Output

When running the proxy and visiting an HTTPS site, you'll see logs like:

```
[INFO] TLS handshake successful for example.com
[INFO] Request: GET https://example.com/
[INFO] Response: 200 OK
[INFO] Body length: 1234 bytes
```


---

## 📦 Certificate Generation Flow

1. On first request to `https://example.com`:
    - Check if certificate is cached
    - If not, generate new cert signed by root CA
    - Store in memory cache
2. Subsequent requests reuse cached cert
3. Certificates are valid for 1 year by default

---

## 🧱 Requirements

- Go 1.20+
- OpenSSL or compatible tooling (optional)
- Root CA installation on client device

---

## 🧪 Known Limitations

- Does not support HTTP/2 or QUIC out of the box
- No UI or web dashboard
- Requires manual root certificate trust setup
- Not suitable for production environments

---

## 🧩 Roadmap / Future Enhancements

| Feature | Status |
|--------|--------|
| WebSocket support | Planned |
| HTTP/2 support | Planned |
| Automatic root CA install (macOS/Linux only) | In Progress |
| GUI interface | TBD |
| Request/Response rewriting UI | TBD |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

## 🤝 Contributions

Contributions are welcome! Please open issues or PRs for:

- Adding UI components
- Improving certificate caching
- Supporting more protocols (e.g., HTTP/2, FTP)
- Adding logging/export features
- Writing integration tests

---

## 📬 Contact

For questions, feedback, or feature requests, feel free to reach out to the maintainers or open an issue.

---

> 📢 **Note:** This software is intended for educational and testing purposes only. Do not use it on networks where you do not