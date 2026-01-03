# Programmatic control of Eero mesh WiFi networks

**Eero has no official public API**, but a well-documented unofficial API powers all community libraries. The Go library **goeero** provides basic functionality, while Python's **eero-client** offers the most mature standalone implementation. For comprehensive control including device blocking, the **Home Assistant integration** provides the most complete feature set—all use the same reverse-engineered REST API at `api-user.e2ro.com`.

## Official API status: none for general use

Eero (owned by Amazon since 2019) does not offer a public API for developers. The company's official stance, confirmed in their community forums: "eero does not have a public API available."

A limited **EU Data Portability API** exists for GDPR compliance, but it requires signing an NDA, completing a security assessment, and is **read-only**—no device control capabilities. This option is impractical for most developers seeking programmatic network management.

The good news: the Eero mobile app uses a REST API that has been thoroughly reverse-engineered since 2016. This unofficial API provides full control capabilities and forms the foundation for all community libraries.

## The unofficial API architecture

All libraries communicate with `https://api-user.e2ro.com` using a consistent pattern. The API uses **cookie-based authentication** with a session token obtained through SMS or email verification—notably, **Amazon account login is not supported**.

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/2.2/login` | POST | Initiates login, returns user_token |
| `/2.2/login/verify` | POST | Verifies OTP code, activates session |
| `/2.2/networks/{id}/devices` | GET | Lists all connected devices |
| `/2.2/networks/{id}/devices/{id}` | PUT | Modify device (block, pause) |
| `/2.2/networks/{id}/profiles` | GET/PUT | Manage family profiles |
| `/2.2/networks/{id}/guestnetwork` | GET/PUT | Guest network control |
| `/2.2/networks/{id}/reboot` | POST | Reboot the network |

The authentication flow requires two steps: POST a phone number or email to `/2.2/login`, receive a verification code via SMS/email, then POST that code to `/2.2/login/verify`. The resulting token persists indefinitely and should be stored locally.

## Go: goeero provides core functionality

**Repository**: https://github.com/ab623/goeero  
**Package**: `github.com/ab623/goeero/eero`  
**License**: GPL-3.0 | **Stars**: 5 | **Latest**: v1.0.0 (August 2023)

This is currently the **only Go library** for Eero control. Installation is straightforward:

```bash
go get github.com/ab623/goeero/eero
```

The library supports authentication, listing networks and devices, and basic account operations:

```go
package main

import (
    "fmt"
    "github.com/ab623/goeero/eero"
)

func main() {
    // With existing token
    client := eero.New(userToken)
    
    // List all devices
    devices, err := client.Devices()
    if err != nil {
        panic(err)
    }
    
    for _, d := range devices {
        fmt.Printf("Device: %s | MAC: %s | IP: %s | Connected: %v\n",
            d.Nickname, d.Mac, d.IP, d.Connected)
    }
}
```

For initial authentication:

```go
// Step 1: Request verification code
loginData, _ := client.Login("+15551234567")

// Step 2: User enters code received via SMS/email
client.LoginVerify(loginData.UserToken, "123456")
```

The library includes a CLI tool for quick testing: `./eeroclient devices | jq` outputs device JSON.

**Limitation**: goeero doesn't implement device blocking or pausing directly. You'd need to extend it with PUT requests to the device endpoint using payload `{"blocked": true}` or `{"paused": true}`.

## Python: the most mature ecosystem

### PyPI package (recommended for new projects)

**Package**: `eero-client` | **Version**: 2.2.0 (August 2024)  
**PyPI**: https://pypi.org/project/eero-client/

```bash
pip install eero-client
```

This modernized library uses Pydantic models and type hints:

```python
from eero import Eero, FileSessionStorage

session = FileSessionStorage("session.cookie")
client = Eero(session=session)

# Authentication (first time only)
if not client.is_authenticated:
    user_token = client.login("+15551234567")
    code = input("Enter verification code: ")
    client.login_verify(code, user_token)

# List connected devices
for device in client.devices:
    print(f"{device.nickname or device.hostname}: {device.ip} ({device.mac})")
    print(f"  Connected: {device.connected}, Paused: {device.paused}")

# View blocked devices
blocked = client.device_blacklist

# Network operations
client.reboot()                    # Reboot entire network
client.reboot_eero(id=eero_id)     # Reboot specific node
client.run_speedtest()             # Trigger speed test
```

Available properties: `account`, `devices`, `device_blacklist`, `eeros`, `guestnetwork`, `profiles`, `speedtest`, `networks`, `diagnostics`, `insights`, `routing`.

### Original library (343max/eero-client)

**GitHub**: https://github.com/343max/eero-client  
**Stars**: 179 | **License**: MIT

The foundational library that all others derive from. Simpler code, useful for learning the API:

```bash
pip install git+https://github.com/343max/eero-client.git
```

## JavaScript: eero-js-api

**GitHub**: https://github.com/bbernstein/eero-js-api  
**NPM**: `eero-js-api` or `eero-node`  
**Stars**: 13 | **License**: MIT

```bash
npm install eero-js-api
```

A direct port of the Python library with identical functionality. Stores session tokens in `/tmp`. Less actively maintained but functional.

## Home Assistant integration: most complete feature set

**GitHub**: https://github.com/schmittx/home-assistant-eero  
**Stars**: 170 | **Latest**: v1.8.1 (September 2025)

While designed for Home Assistant, this integration's source code serves as the **best API documentation available**—it implements features no standalone library offers:

- **Device blocking and pausing** (clients and profiles)
- **Content filters** per profile
- **Guest network** enable/disable with settings
- **App blocking** per profile (requires Eero Secure subscription)
- **Activity data** and bandwidth metrics (requires Eero Plus)
- **QR code generation** for network joining
- **Firmware update control**
- **Nightlight control** for Eero Beacon devices

For developers building Go or Python applications, examining this codebase reveals the complete API surface area for blocking devices:

```python
# Conceptual example based on HA integration patterns
# PUT /2.2/networks/{network_id}/devices/{device_id}
payload = {"paused": True}   # Pause internet access
payload = {"blocked": True}  # Block device entirely
```

Install via HACS (Home Assistant Community Store) by adding as a custom repository.

## Authentication implementation details

The critical insight: **you must use email or phone authentication, not Amazon login**. Users who set up Eero with Amazon credentials need to create a secondary admin account.

**Workaround for Amazon users**: In the Eero app, invite a new email address as a network admin. Use that email for API authentication.

Session tokens persist indefinitely once verified. Store them in a file:

```python
# Python pattern for persistent sessions
class FileSessionStorage:
    def __init__(self, path):
        self.path = path
        try:
            with open(path, 'r') as f:
                self._token = f.read().strip()
        except FileNotFoundError:
            self._token = None
    
    @property
    def cookie(self):
        return self._token
    
    @cookie.setter
    def cookie(self, value):
        self._token = value
        with open(self.path, 'w') as f:
            f.write(value)
```

Required HTTP headers for all requests:
```
User-Agent: eero-ios/2.16.0 (iPhone8,1; iOS 11.3)
Cookie: s=[USER_TOKEN]
Content-Type: application/json
```

## Rate limits and restrictions

Eero hasn't published official rate limits, but community experience suggests:

- **Minimum polling interval**: 25 seconds (used by Home Assistant integration)
- **Account data caching**: 1 hour recommended
- **No reported bans** for normal usage patterns
- **Aggressive polling may trigger throttling** (unconfirmed)

**Terms of Service consideration**: Using the unofficial API technically violates Eero's ToS prohibition on reverse engineering. However, thousands of users have run these integrations since 2016 without enforcement action.

## C/C++ status

**No C/C++ library exists**. The API is straightforward REST/JSON, so creating one using libcurl would be feasible using the endpoint documentation above.

## Recommendations by use case

| Goal | Best Option |
|------|-------------|
| Go application with basic features | **goeero** (extend for blocking) |
| Python application | **pip install eero-client** |
| Full device blocking/pausing | Reference **Home Assistant integration** source |
| Learning the API | **343max/eero-client** (clean, readable) |
| Production Home Assistant | **schmittx/home-assistant-eero** via HACS |
| Prometheus monitoring | **brmurphy/eero-exporter** |

## Conclusion

The Eero ecosystem lacks official API support, but the community has built robust tooling around the reverse-engineered mobile app API. **For Go development**, goeero provides a foundation that you'd need to extend for blocking features. **For Python**, the PyPI eero-client package offers the cleanest standalone experience. The **Home Assistant integration** remains the most complete reference implementation—its source code documents API endpoints that no other library exposes, making it invaluable for developers building comprehensive Eero control applications.