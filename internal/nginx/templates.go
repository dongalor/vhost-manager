package nginx

const defaultTemplate = `server {
    listen {{.Port}};
    server_name {{.ServerName}};

    root {{.DocumentRoot}};
    index index.html index.htm index.php;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss;

    location / {
        try_files $uri $uri/ =404;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
    }

    # Deny access to backup and config files
    location ~* \.(bak|config|dist|fla|inc|ini|log|sh|sql|tar|gz)$ {
        deny all;
    }

    # Cache static assets
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # PHP support (if needed)
    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
        fastcgi_index index.php;
        include fastcgi_params;
    }
}`

const redirectHTTPSTemplate = `# Redirect HTTP to HTTPS
server {
    listen {{.Port}};
    server_name {{.ServerName}};
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name {{.ServerName}};

    root {{.DocumentRoot}};
    index index.html index.htm index.php;

    # SSL configuration (to be configured by certbot)
    # ssl_certificate /etc/letsencrypt/live/{{.ServerName}}/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/{{.ServerName}}/privkey.pem;
    # include /etc/letsencrypt/options-ssl-nginx.conf;
    # ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss;

    location / {
        try_files $uri $uri/ =404;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
    }

    # Deny access to backup and config files
    location ~* \.(bak|config|dist|fla|inc|ini|log|sh|sql|tar|gz)$ {
        deny all;
    }

    # Cache static assets
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # PHP support (if needed)
    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
        fastcgi_index index.php;
        include fastcgi_params;
    }
}`

const basicTemplate = `server {
    listen {{.Port}};
    server_name {{.ServerName}};

    root {{.DocumentRoot}};
    index index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }
}`