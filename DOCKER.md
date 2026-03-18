# Docker Deployment Guide

This guide explains how to run YouClips using Docker Compose.

## Prerequisites

- Docker Engine 20.10+ 
- Docker Compose 2.0+
- At least 2GB free disk space

## Quick Start

1. **Clone the repository** (if you haven't already):
   ```bash
   git clone <your-repo>
   cd YouClips
   ```

2. **Create data directory**:
   ```bash
   mkdir -p data
   ```

3. **Build and start services**:
   ```bash
   docker-compose up -d
   ```

4. **Access the application**:
   - Frontend: http://localhost:3000
   - API: http://localhost:8080

## Configuration

### Environment Variables

You can customize the behavior by editing `docker-compose.yml`:

#### API Service
- `DB_PATH`: Database file path (default: `/app/data/youclips.db`)
- `STORAGE_DIR`: Temporary clips storage (default: `/app/storage/clips`)
- `CLIP_TTL_MINUTES`: How long clips stay available before auto-deletion (default: `5`)

#### Web Service
- `ORIGIN`: Frontend URL (default: `http://localhost:3000`)
- `API_URL`: Backend API URL (default: `http://api:8080`)

### Custom Configuration Example

```yaml
services:
  api:
    environment:
      - CLIP_TTL_MINUTES=60  # Keep clips for 1 hour
```

## Docker Compose Commands

### Start services
```bash
docker-compose up -d
```

### View logs
```bash
# All services
docker-compose logs -f

# API only
docker-compose logs -f api

# Web only
docker-compose logs -f web
```

### Stop services
```bash
docker-compose down
```

### Stop and remove volumes (⚠️ deletes database)
```bash
docker-compose down -v
```

### Rebuild images
```bash
docker-compose build --no-cache
docker-compose up -d
```

### Check service status
```bash
docker-compose ps
```

## Data Persistence

- **Database**: Stored in `./data/youclips.db` (persisted on host)
- **Clips**: Stored in Docker volume `clips-storage` (auto-cleaned by TTL)
- **Metadata cache**: Part of database (permanent)

## Troubleshooting

### API not starting
Check logs:
```bash
docker-compose logs api
```

Common issues:
- Port 8080 already in use
- Missing yt-dlp or ffmpeg (should be in Docker image)

### Frontend not connecting to API
Check:
1. API service is running: `docker-compose ps api`
2. API health: `curl http://localhost:8080/clips`
3. Frontend logs: `docker-compose logs web`

### Out of disk space
Clips are auto-deleted after TTL, but you can manually clean:
```bash
# Remove old clips volume
docker-compose down
docker volume rm youclips_clips-storage
docker-compose up -d
```

## Production Deployment

For production, consider:

1. **Change ports** in `docker-compose.yml`:
   ```yaml
   services:
     web:
       ports:
         - "80:3000"  # Use port 80
   ```

2. **Increase TTL** for better UX:
   ```yaml
   environment:
     - CLIP_TTL_MINUTES=60  # 1 hour
   ```

3. **Add reverse proxy** (nginx/traefik) for HTTPS

4. **Resource limits**:
   ```yaml
   services:
     api:
       deploy:
         resources:
           limits:
             cpus: '2'
             memory: 2G
   ```

5. **Backup database** regularly:
   ```bash
   docker cp youclips-api:/app/data/youclips.db ./backup/
   ```

## Updating

To update to latest version:

```bash
git pull
docker-compose build --no-cache
docker-compose up -d
```

## Monitoring

Health checks are configured for both services:
- API: http://localhost:8080/clips
- Web: http://localhost:3000

Check health status:
```bash
docker-compose ps
```

## Support

For issues, check:
1. Service logs: `docker-compose logs`
2. Container status: `docker-compose ps`
3. Disk space: `df -h`
