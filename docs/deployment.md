# Deployment Guide

Complete guide for deploying Home Agent to production.

## Prerequisites

- Linux server (Ubuntu 20.04+ recommended)
- Domain name configured
- SSL certificate (Let's Encrypt recommended)
- Go 1.21+ installed
- Node.js 18+ installed (for building)
- Nginx installed
- Systemd (for service management)

## Deployment Architecture

```
Internet
   ↓
Nginx (Port 443 HTTPS)
   ↓
Go Backend (Port 8080)
   ↓
Claude API
```

## Step-by-Step Deployment

### 1. Server Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y nginx certbot python3-certbot-nginx git golang-go nodejs npm

# Verify versions
go version    # Should be 1.21+
node --version # Should be 18+
```

### 2. Create Deployment User

```bash
# Create dedicated user
sudo useradd -r -s /bin/bash -d /opt/home-agent homeagent
sudo mkdir -p /opt/home-agent
sudo chown homeagent:homeagent /opt/home-agent
```

### 3. Clone and Build

```bash
# Switch to deployment user
sudo su - homeagent

# Clone repository
cd /opt/home-agent
git clone https://github.com/yourusername/home-agent.git
cd home-agent

# Build frontend
cd frontend
npm install
npm run build

# Build backend
cd ../backend
go mod download
go build -ldflags="-s -w" -o home-agent

# Exit deployment user
exit
```

### 4. Configure Environment

```bash
# Create environment file
sudo nano /etc/home-agent/.env
```

Add configuration:
```env
ANTHROPIC_API_KEY=sk-ant-your-key-here
PORT=8080
HOST=localhost
DB_PATH=/opt/home-agent/data/home-agent.db
```

Secure the file:
```bash
sudo chown homeagent:homeagent /etc/home-agent/.env
sudo chmod 600 /etc/home-agent/.env
```

### 5. Setup Systemd Service

```bash
# Copy service file
sudo cp /opt/home-agent/home-agent/home-agent.service.example /etc/systemd/system/home-agent.service

# Edit service file
sudo nano /etc/systemd/system/home-agent.service
```

Update paths in the service file:
```ini
WorkingDirectory=/opt/home-agent/home-agent/backend
ExecStart=/opt/home-agent/home-agent/backend/home-agent
EnvironmentFile=/etc/home-agent/.env
```

Enable and start service:
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable home-agent

# Start service
sudo systemctl start home-agent

# Check status
sudo systemctl status home-agent
```

### 6. Configure Nginx

```bash
# Copy nginx configuration
sudo cp /opt/home-agent/home-agent/nginx.conf.example /etc/nginx/sites-available/home-agent

# Edit configuration
sudo nano /etc/nginx/sites-available/home-agent
```

Update domain name:
```nginx
server_name yourdomain.com www.yourdomain.com;
```

Enable site:
```bash
# Create symlink
sudo ln -s /etc/nginx/sites-available/home-agent /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

### 7. Setup SSL Certificate

```bash
# Get certificate with certbot
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Verify auto-renewal
sudo certbot renew --dry-run
```

### 8. Configure Firewall

```bash
# Allow SSH, HTTP, HTTPS
sudo ufw allow 22
sudo ufw allow 80
sudo ufw allow 443

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

### 9. Verify Deployment

```bash
# Check backend service
sudo systemctl status home-agent

# Check nginx
sudo systemctl status nginx

# Test WebSocket connection
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost:8080/ws

# Check logs
sudo journalctl -u home-agent -f
```

### 10. Access Application

Open browser to: `https://yourdomain.com`

## Monitoring

### View Logs

```bash
# Backend logs
sudo journalctl -u home-agent -f

# Nginx access logs
sudo tail -f /var/log/nginx/home-agent-access.log

# Nginx error logs
sudo tail -f /var/log/nginx/home-agent-error.log
```

### System Status

```bash
# Service status
sudo systemctl status home-agent

# Resource usage
htop

# Disk usage
df -h

# Network connections
netstat -tulpn | grep 8080
```

## Maintenance

### Update Application

```bash
# Switch to deployment user
sudo su - homeagent

# Pull latest code
cd /opt/home-agent/home-agent
git pull

# Rebuild frontend
cd frontend
npm install
npm run build

# Rebuild backend
cd ../backend
go build -ldflags="-s -w" -o home-agent

# Exit deployment user
exit

# Restart service
sudo systemctl restart home-agent
```

### Backup

```bash
# Backup database
sudo -u homeagent cp /opt/home-agent/data/home-agent.db /opt/home-agent/backups/home-agent-$(date +%Y%m%d).db

# Backup configuration
sudo cp /etc/home-agent/.env /opt/home-agent/backups/.env-$(date +%Y%m%d)
```

### Restore

```bash
# Restore database
sudo systemctl stop home-agent
sudo -u homeagent cp /opt/home-agent/backups/home-agent-20240101.db /opt/home-agent/data/home-agent.db
sudo systemctl start home-agent
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u home-agent -n 50

# Check configuration
sudo -u homeagent /opt/home-agent/home-agent/backend/home-agent

# Check permissions
ls -la /opt/home-agent/home-agent/backend/home-agent
```

### WebSocket Connection Fails

```bash
# Check nginx configuration
sudo nginx -t

# Check backend is listening
netstat -tulpn | grep 8080

# Check firewall
sudo ufw status

# Test directly
wscat -c ws://localhost:8080/ws
```

### High Memory Usage

```bash
# Check process
ps aux | grep home-agent

# Restart service
sudo systemctl restart home-agent

# Check for leaks
sudo journalctl -u home-agent | grep -i "memory"
```

### SSL Certificate Issues

```bash
# Check certificate
sudo certbot certificates

# Renew certificate
sudo certbot renew

# Reload nginx
sudo systemctl reload nginx
```

## Security Hardening

### 1. Firewall Rules

```bash
# Allow only necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```

### 2. Fail2Ban

```bash
# Install fail2ban
sudo apt install -y fail2ban

# Configure
sudo nano /etc/fail2ban/jail.local
```

Add:
```ini
[nginx-http-auth]
enabled = true

[nginx-noscript]
enabled = true
```

Start service:
```bash
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 3. Regular Updates

```bash
# Setup automatic updates
sudo apt install -y unattended-upgrades

# Configure
sudo dpkg-reconfigure -plow unattended-upgrades
```

### 4. Monitoring

```bash
# Setup monitoring (optional)
# - Prometheus
# - Grafana
# - Alertmanager
```

## Performance Optimization

### 1. Nginx Tuning

Edit `/etc/nginx/nginx.conf`:
```nginx
worker_processes auto;
worker_connections 1024;

# Enable gzip
gzip on;
gzip_vary on;
gzip_types text/plain text/css application/json application/javascript;

# Buffer settings
client_body_buffer_size 10K;
client_header_buffer_size 1k;
client_max_body_size 8m;
large_client_header_buffers 2 1k;
```

### 2. System Limits

Edit `/etc/security/limits.conf`:
```
homeagent soft nofile 65535
homeagent hard nofile 65535
```

### 3. Backend Optimization

Set environment variables:
```bash
GOMAXPROCS=4  # Number of CPU cores
```

## Scaling

### Horizontal Scaling

1. Deploy multiple backend instances
2. Setup load balancer (HAProxy/Nginx)
3. Use Redis for shared sessions
4. Configure sticky sessions for WebSocket

### Vertical Scaling

1. Increase server resources
2. Optimize database queries
3. Enable caching
4. Use CDN for static assets

## Disaster Recovery

### Automated Backups

```bash
# Create backup script
sudo nano /opt/home-agent/scripts/backup.sh
```

Add:
```bash
#!/bin/bash
BACKUP_DIR="/opt/home-agent/backups"
DATE=$(date +%Y%m%d-%H%M%S)

# Backup database
cp /opt/home-agent/data/home-agent.db "$BACKUP_DIR/db-$DATE.db"

# Backup config
cp /etc/home-agent/.env "$BACKUP_DIR/env-$DATE"

# Remove old backups (keep last 7 days)
find "$BACKUP_DIR" -type f -mtime +7 -delete
```

Setup cron:
```bash
sudo crontab -e
```

Add:
```
0 2 * * * /opt/home-agent/scripts/backup.sh
```

## Rollback Procedure

```bash
# Stop service
sudo systemctl stop home-agent

# Revert code
cd /opt/home-agent/home-agent
git checkout <previous-commit>

# Rebuild
cd frontend && npm run build
cd ../backend && go build -o home-agent

# Restore database
cp /opt/home-agent/backups/db-latest.db /opt/home-agent/data/home-agent.db

# Start service
sudo systemctl start home-agent
```

## Health Checks

### Automated Monitoring

Create health check endpoint and monitor with:
- UptimeRobot
- Pingdom
- Datadog
- Custom monitoring script

### Manual Checks

```bash
# Check service
curl https://yourdomain.com

# Check WebSocket
wscat -c wss://yourdomain.com/ws

# Check SSL
openssl s_client -connect yourdomain.com:443

# Check response time
curl -w "@curl-format.txt" -o /dev/null -s https://yourdomain.com
```

## Support

For deployment issues:
1. Check logs: `sudo journalctl -u home-agent`
2. Review nginx logs: `/var/log/nginx/`
3. Test components individually
4. Open issue on GitHub

## Checklist

Pre-deployment:
- [ ] Server provisioned
- [ ] Domain configured
- [ ] SSL certificate obtained
- [ ] Environment variables set
- [ ] Firewall configured

Deployment:
- [ ] Code deployed
- [ ] Frontend built
- [ ] Backend built
- [ ] Service installed
- [ ] Nginx configured
- [ ] SSL configured

Post-deployment:
- [ ] Application accessible
- [ ] WebSocket working
- [ ] Logs being generated
- [ ] Backups configured
- [ ] Monitoring setup
- [ ] Documentation updated

## Resources

- [Let's Encrypt](https://letsencrypt.org/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Systemd Documentation](https://systemd.io/)
- [Go Deployment Best Practices](https://golang.org/doc/)

---

End of Deployment Guide
