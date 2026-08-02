# Production SSL/TLS Configuration (Optional)

# To enable HTTPS in production:
# 1. Generate or obtain SSL certificates
# 2. Place them in nginx/certs/ directory
# 3. Uncomment and configure the HTTPS sections in nginx.conf

# Example using Let's Encrypt with Certbot:
# docker run --rm -it -v /path/to/certs:/etc/letsencrypt certbot/certbot certonly --standalone -d yourdomain.com

# Once you have certificates, update nginx.conf with:
# - Uncomment the HTTP to HTTPS redirect section
# - Add HTTPS server block with ssl_certificate and ssl_certificate_key
# - Configure SSL parameters (protocols, ciphers, etc.)
