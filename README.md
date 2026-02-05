# Monstache-Milvus

[![Go Version](https://img.shields.io/github/go-mod/go-version/doing-cr7/monstache-milvus)](https://go.dev/)
[![License](https://img.shields.io/github/license/doing-cr7/monstache-milvus)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/doing-cr7/monstache-milvus)](https://github.com/doing-cr7/monstache-milvus/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/doing-cr7/monstache-milvus)](https://github.com/doing-cr7/monstache-milvus/issues)

> A Go daemon that syncs MongoDB to Elasticsearch **and Milvus** in realtime. Perfect for building hybrid search systems combining traditional text search with vector similarity search.

**Based on [Monstache](https://github.com/rwynn/monstache) by Ryan Wynn, with added support for [Milvus](https://milvus.io/) vector database.**

---

## ✨ Features

### 🎯 Core Capabilities
- **Dual-Engine Sync**: Simultaneously sync MongoDB to both Elasticsearch and Milvus
- **Real-time Streaming**: Uses MongoDB Change Streams for instant data updates
- **Vector Search Ready**: Native support for Milvus/Zilliz Cloud vector database
- **High Availability**: Cluster mode with automatic failover
- **Direct Reads**: Bulk load existing data with parallel processing

### 🚀 Advanced Features
- **Flexible Mapping**: Custom field mappings and transformations via JavaScript or Go plugins
- **Data Relations**: Support for document relationships across collections
- **GridFS Support**: Index file content from MongoDB GridFS
- **Time Machine**: Historical data indexing with timestamps
- **Alerting**: Built-in alert system (Feishu/Lark integration, customizable)
- **Monitoring**: HTTP endpoints for health checks and statistics

### 🔧 Production Ready
- Resume from last position (timestamp or token-based)
- Configurable batch sizes and concurrency
- Comprehensive logging and metrics
- Docker and Kubernetes support
- Automatic reconnection and error handling

---

## 📋 Table of Contents

- [Why Monstache-Milvus?](#-why-monstache-milvus)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage Examples](#-usage-examples)
- [MongoDB Setup](#-mongodb-setup)
- [Development](#-development)
- [FAQ](#-faq)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🎯 Why Monstache-Milvus?

### Use Cases

**Perfect for building:**
- 🔍 **Hybrid Search Systems**: Combine keyword search (Elasticsearch) with semantic search (Milvus)
- 🤖 **AI/ML Applications**: Sync embeddings from MongoDB to Milvus for similarity search
- 📊 **Real-time Analytics**: Keep your search and vector databases in sync with MongoDB
- 🔄 **Data Migration**: Migrate large MongoDB datasets to Elasticsearch and Milvus efficiently

### Comparison with Original Monstache

| Feature | Original Monstache | Monstache-Milvus |
|---------|-------------------|------------------|
| Elasticsearch Sync | ✅ | ✅ |
| Milvus/Zilliz Sync | ❌ | ✅ |
| Dual Engine Writes | ❌ | ✅ |

---

## 🏗️ Architecture

```mermaid
graph TB
    MongoDB[(MongoDB)] -->|Change Streams| Monstache[Monstache-Milvus]
    MongoDB -->|Direct Read| Monstache
    
    Monstache -->|Text Data| Elasticsearch[(Elasticsearch)]
    Monstache -->|Vector Data| Milvus[(Milvus/Zilliz)]
    
    Monstache -->|GridFS| Files[File Processing]
    Monstache -->|Scripts| Transform[JS/Go Plugins]
    
    style Monstache fill:#4CAF50
    style MongoDB fill:#47A248
    style Elasticsearch fill:#005571
    style Milvus fill:#00ADD8
```

**Detailed architecture diagram:** [architecture.mermaid](architecture.mermaid)

### Data Flow

1. **Change Detection**: Monitors MongoDB using Change Streams or Oplog
2. **Transformation**: Apply custom mappings, filters, and transformations
3. **Dual Write**: 
   - Milvus receives vector data for similarity search
   - Elasticsearch receives full documents for text search
4. **Progress Tracking**: Save resume tokens for fault tolerance

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (for building from source)
- MongoDB 3.6+ (4.0+ recommended for Change Streams)
- Elasticsearch 7.0+ (optional)
- Milvus 2.0+ or Zilliz Cloud account (optional)

### One-Minute Setup

```bash
# 1. Clone the repository
git clone https://github.com/doing-cr7/monstache-milvus.git
cd monstache-milvus

# 2. Copy and configure
cp config.example.toml config.toml
vim config.toml  # Edit with your MongoDB, ES, and Milvus credentials

# 3. Build and run
make build
./bin/monstache -f config.toml
```

### Basic Configuration

```toml
# MongoDB connection
mongo-url = "mongodb://user:pass@localhost:27017"

# Elasticsearch (optional)
elasticsearch-urls = ["http://localhost:9200"]

# Milvus/Zilliz (optional)
zilliz-enabled = true
zilliz-addr = "https://your-cluster.zillizcloud.com:19530"
zilliz-api-key = "your-api-key"
zilliz-collection-name = "your_collection"

# What to sync
change-stream-namespaces = ["mydb.mycollection"]
```

### Verify It's Working

```bash
# Check health
curl http://localhost:8080/healthz

# Check statistics
curl http://localhost:8080/stats
```

---

## 📦 Installation

### Option 1: Build from Source

```bash
# Clone repository
git clone https://github.com/doing-cr7/monstache-milvus.git
cd monstache-milvus

# Build binary
go build -o bin/monstache monstache.go

# Or use Makefile
make build
```

### Option 2: Docker

```bash
# Using Docker
docker pull doing-cr7/monstache-milvus:latest

docker run -d \
  -v /path/to/config.toml:/config.toml \
  doing-cr7/monstache-milvus:latest \
  -f /config.toml
```

### Option 3: Docker Compose

```yaml
version: '3.8'
services:
  monstache:
    image: doing-cr7/monstache-milvus:latest
    volumes:
      - ./config.toml:/config.toml
    command: -f /config.toml
    environment:
      - MONSTACHE_MONGO_URL=${MONGO_URL}
      - MONSTACHE_ZILLIZ_API_KEY=${ZILLIZ_API_KEY}
    restart: unless-stopped
```

### Option 4: Kubernetes

See [docker/release/README.md](docker/release/README.md) for Kubernetes deployment examples.

---

## ⚙️ Configuration

### Minimal Configuration

```toml
mongo-url = "mongodb://localhost:27017"
elasticsearch-urls = ["http://localhost:9200"]
change-stream-namespaces = [""]  # Watch all databases
```

### Production Configuration

```toml
# MongoDB
mongo-url = "mongodb://user:pass@mongo1:27017,mongo2:27017/admin?replicaSet=rs0"

# Elasticsearch
elasticsearch-urls = ["http://es1:9200", "http://es2:9200"]
elasticsearch-max-conns = 10

# Milvus
zilliz-enabled = true
zilliz-addr = "your-milvus-endpoint:19530"
zilliz-api-key = "your-api-key"
zilliz-collection-name = "embeddings"
zilliz-max-conns = 4
zilliz-max-docs = 256

# High Availability
cluster-name = "prod-sync-cluster"
resume = true
resume-strategy = 1  # Token-based

# Performance
direct-read-concur = 4
elasticsearch-max-docs = 1000

# Monitoring
enable-http-server = true
http-server-addr = ":8080"
```

### Configuration File Examples

- **Basic**: [config.example.toml](config.example.toml)
- **Detailed Guide**: [CONFIGURATION.md](CONFIGURATION.md)
- **Environment Variables**: See [CONFIGURATION.md](CONFIGURATION.md#environment-variables)

---

## 💡 Usage Examples

### Example 1: Sync MongoDB to Elasticsearch

```toml
mongo-url = "mongodb://localhost:27017"
elasticsearch-urls = ["http://localhost:9200"]
change-stream-namespaces = ["mydb.products"]

[[mapping]]
namespace = "mydb.products"
index = "products_index"
```

### Example 2: Sync Vectors to Milvus

```toml
mongo-url = "mongodb://localhost:27017"

# Enable Milvus sync
zilliz-enabled = true
zilliz-addr = "localhost:19530"
zilliz-api-key = "your-key"
zilliz-collection-name = "document_embeddings"

# Sync specific collection with embeddings
change-stream-namespaces = ["mydb.documents"]
```

### Example 3: Dual Engine Sync (Hybrid Search)

```toml
# Sync to both Elasticsearch and Milvus
mongo-url = "mongodb://localhost:27017"

# Text search in Elasticsearch
elasticsearch-urls = ["http://localhost:9200"]

# Vector search in Milvus
zilliz-enabled = true
zilliz-addr = "localhost:19530"
zilliz-api-key = "your-key"
zilliz-collection-name = "vectors"

# Watch same collection
change-stream-namespaces = ["mydb.articles"]
```

### Example 4: Custom Transformations

Create a JavaScript transformation:

```javascript
// transform.js
module.exports = function(doc) {
  // Add computed field
  doc.fullName = doc.firstName + " " + doc.lastName;
  
  // Filter out sensitive data
  delete doc.password;
  
  return doc;
}
```

Configure it:

```toml
[[script]]
namespace = "mydb.users"
path = "./transform.js"
```

### Example 5: Direct Read (Initial Sync)

```toml
# Bulk load existing data
direct-read-namespaces = ["mydb.products"]
direct-read-concur = 4  # Parallel workers
direct-read-split-max = 4  # Split large collections

# Exit after initial sync (optional)
exit-after-direct-reads = true
```

---

## 🔐 MongoDB Setup

### Required Permissions

Monstache requires specific MongoDB permissions to function properly.

#### Option 1: Minimal Permissions (Recommended)

```javascript
// Connect to MongoDB
use admin

// Create dedicated user
db.createUser({
  user: "",
  pwd: "",
  roles: [
    { role: "readWrite", db: "admin" },
    { role: "readWrite", db: "<logic db>" },
    { role: "readWrite", db: "monstache" },
    { role: "clusterMonitor", db: "admin" }
  ]
})
```



### Connection String

```bash
# Replica Set (recommended)
mongodb://monstache:password@mongo1:27017,mongo2:27017/?replicaSet=rs0

# Standalone (for development only)
mongodb://monstache:password@localhost:27017

# With authentication database
mongodb://monstache:password@localhost:27017/admin?authSource=admin
```


---

## 👨‍💻 Development

### Building from Source

```bash
# Install dependencies
go mod download

# Build
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 make build
```


### Project Structure

```
monstache-milvus/
├── monstache.go           # Main application
├── monstache_test.go      # Tests
├── dao/
│   └── milvus/           # Milvus integration
├── pkg/
│   └── oplog/            # Oplog processing
├── monstachemap/         # Plugin system
├── docker/               # Docker configurations
└── config.example.toml   # Example configuration
```



**Steps to contribute:**
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## ❓ FAQ

### General Questions

**Q: What's the difference from original Monstache?**  
A: We've added native Milvus/Zilliz support for vector search, enabling hybrid search systems that combine text and semantic search.

**Q: Can I sync to only Milvus (without Elasticsearch)?**  
A: Yes! Set `zilliz-enabled = true` and omit `elasticsearch-urls`. You can use either or both.

**Q: Does it support MongoDB standalone?**  
A: For development, yes. For production, MongoDB replica set is required for Change Streams.

**Q: What happens if Monstache crashes?**  
A: It resumes from the last saved position (timestamp or token) when `resume = true` is configured.

### Performance

**Q: How fast is the sync?**  
A: Depends on your setup. Typically processes 1000-5000 docs/sec. Use `elasticsearch-max-conns` and `zilliz-max-conns` to tune.

**Q: How to handle large existing datasets?**  
A: Use `direct-read-namespaces` with `direct-read-concur` for parallel bulk loading.

### Troubleshooting

**Q: "Unable to connect to MongoDB"**  
A: Check connection string, replica set name, and network connectivity. Verify with `mongo` CLI first.

**Q: "Change streams are not supported"**  
A: Requires MongoDB 3.6+ in replica set mode. For standalone, use `enable-oplog = true` (legacy).

**Q: "Zilliz collection not found"**  
A: Create the collection in Milvus/Zilliz first. Monstache doesn't auto-create collections.

**Q: Performance is slow**  
A: Tune batch sizes (`elasticsearch-max-docs`, `zilliz-max-docs`), increase workers (`elasticsearch-max-conns`), or check network latency.

**For more issues:** [GitHub Issues](https://github.com/doing-cr7/monstache-milvus/issues)



### Monitoring & Alerting

```toml
# Enable HTTP server
enable-http-server = true
http-server-addr = ":8080"

# Feishu/Lark alerts (customize for your system)
is-feishu = true
alert-api-url = "https://your-webhook-url"
alert-robot-key = "your-key"
```

Endpoints:
- `GET /healthz` - Health check
- `GET /stats` - Sync statistics
- `GET /instance` - Instance information

---

## 📊 Performance Tuning

### Elasticsearch Optimization

```toml
elasticsearch-max-conns = 10      # Concurrent workers
elasticsearch-max-docs = 1000     # Batch size
elasticsearch-max-bytes = 8388608 # 8MB batch size
elasticsearch-max-seconds = 1     # Flush interval
```

### Milvus Optimization

```toml
zilliz-max-conns = 4         # Concurrent workers
zilliz-max-docs = 256        # Batch size
zilliz-max-bytes = 2097152   # 2MB batch size
zilliz-max-seconds = 500     # 0.5s flush interval (in ms)
```

### Direct Read Performance

```toml
direct-read-concur = 4      # Parallel workers
direct-read-split-max = 4   # Split large collections
direct-read-no-timeout = true  # No cursor timeout
```

---

## 🙏 Acknowledgments

This project is built upon the excellent work of:
- **[Monstache Project](https://github.com/rwynn/monstache)** - The foundation of this tool
- **[Milvus Team](https://milvus.io/)** - Amazing vector database

Special thanks to all contributors who help improve this project!

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

### Third-Party Licenses

This project uses:
- [Monstache](https://github.com/rwynn/monstache) - MIT License
- [Milvus Go SDK](https://github.com/milvus-io/milvus-sdk-go) - Apache 2.0 License
- [Elastic Go Client](https://github.com/olivere/elastic) - MIT License
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - Apache 2.0 License

---

## 📬 Contact & Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/doing-cr7/monstache-milvus/issues)
- 💡 **Feature Requests**: [GitHub Issues](https://github.com/doing-cr7/monstache-milvus/issues)
---

## ⭐ Star History

<picture>
  <source
    media="(prefers-color-scheme: dark)"
    srcset="https://api.star-history.com/svg?repos=doing-cr7/monstache-milvus&type=Date&theme=dark"
  />
  <source
    media="(prefers-color-scheme: light)"
    srcset="https://api.star-history.com/svg?repos=doing-cr7/monstache-milvus&type=Date"
  />
  <img
    alt="Star History Chart"
    src="https://api.star-history.com/svg?repos=doing-cr7/monstache-milvus&type=Date"
  />
</picture>

---

<div align="center">

**If you find this project helpful, please consider giving it a ⭐!**

Made with ❤️ by the Monstache-Milvus community

[Documentation](CONFIGURATION.md) • [Issues](https://github.com/doing-cr7/monstache-milvus/issues) • [Contributing](CONTRIBUTING.md)

</div>
