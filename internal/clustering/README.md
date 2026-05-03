# Interaction Clustering and Noise Compression

This package provides interaction clustering and noise compression capabilities for GODNSLOG.

## Features

- **Interaction Clustering**: Groups similar interactions together based on type, source IP, token, and patterns
- **Noise Detection**: Identifies and marks high-frequency or known noise patterns
- **Data Compression**: Reduces interaction data size by truncating raw data and compressing headers
- **Duplicate Removal**: Removes duplicate interactions to reduce storage overhead

## Components

### Clusterer

Groups interactions into clusters based on similarity:

```go
clusterer := clustering.NewClusterer(nil)
clusters := clusterer.ClusterInteractions(interactions)
```

**Clustering Strategy**:
- Key: Type + SourceIP + Token
- Pattern extraction: Domain (DNS) or Path (HTTP)
- Noise detection: High frequency or known patterns

### Compressor

Reduces interaction data size:

```go
compressor := clustering.NewCompressor(nil)
compressed := compressor.CompressInteractions(interactions)
```

**Compression Strategy**:
- Truncate raw data to max length (default: 1024)
- Compress headers (keep only important headers)
- Remove duplicate interactions
- Keep first N interactions per cluster

## Configuration

### Clustering Configuration

```go
config := &clustering.ClusteringConfig{
    MaxClusterSize: 100,
    TimeWindow:     "5m",
    NoiseThreshold: 10,
    NoisePatterns: []clustering.NoisePattern{
        {
            Type:        "dns",
            Pattern:     `.*\.google\.com$`,
            Description: "Google DNS queries",
        },
    },
}
```

### Compression Configuration

```go
config := &clustering.CompressionConfig{
    MaxRawDataLength: 1024,
    CompressHeaders:  true,
    RemoveDuplicates: true,
    KeepFirstN:       10,
}
```

## API Endpoints

### POST /clustering/cluster

Cluster interactions and detect noise.

**Request**:
```json
{
  "interactions": [...]
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "clusters": [...],
    "total_count": 10,
    "noise_count": 3
  }
}
```

### POST /clustering/compress

Compress interaction data.

**Request**:
```json
{
  "interactions": [...]
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": [...]
}
```

### GET /clustering/config

Get current configuration.

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "clustering": {...},
    "compression": {...}
  }
}
```

## Integration

### With Rule Engine

Use clustering to reduce noise before rule evaluation:

```go
// Cluster interactions first
clusterer := clustering.NewClusterer(nil)
clusters := clusterer.ClusterInteractions(interactions)

// Filter out noise
var filtered []interaction.Interaction
for _, cluster := range clusters {
    if !cluster.IsNoise {
        filtered = append(filtered, cluster.Interactions...)
    }
}

// Evaluate rules on filtered interactions
engine.Evaluate(ctx, filtered...)
```

### With Storage

Compress interactions before storage:

```go
compressor := clustering.NewCompressor(nil)
compressed := compressor.CompressInteractions(interactions)

// Store compressed interactions
store.SaveInteractions(ctx, compressed)
```

## Noise Patterns

Default noise patterns include:

- Google DNS queries (`.google.com`)
- Cloudflare DNS queries (`.cloudflare.com`)
- Favicon requests (`favicon.ico`)
- Robots.txt requests (`robots.txt`)

Add custom patterns in configuration:

```go
config.NoisePatterns = append(config.NoisePatterns, clustering.NoisePattern{
    Type:        "dns",
    Pattern:     `.*\.internal\.company\.com$`,
    Description: "Internal company DNS",
})
```

## Performance Considerations

- Clustering is O(n) where n is the number of interactions
- Compression reduces storage requirements significantly
- Noise detection reduces false positives in rule evaluation
- Configure time window to balance real-time detection vs. accuracy

## Best Practices

1. **Adjust Noise Threshold**: Set appropriate threshold based on your traffic patterns
2. **Custom Noise Patterns**: Add patterns specific to your environment
3. **Compression Settings**: Balance data retention vs. storage cost
4. **Regular Review**: Periodically review noise patterns and adjust
5. **Monitor Compression Ratio**: Track compression effectiveness
