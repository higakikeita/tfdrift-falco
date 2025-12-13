# Custom Domain Setup for Vercel

This guide explains how to configure a custom domain for your TFDrift-Falco website hosted on Vercel.

## Prerequisites

- A Vercel account with your project deployed
- A domain name purchased from a domain registrar (e.g., Namecheap, GoDaddy, Google Domains, etc.)

## Step 1: Add Domain in Vercel

1. Go to your project in the [Vercel Dashboard](https://vercel.com/dashboard)
2. Click on **Settings** tab
3. Navigate to **Domains** section
4. Click **Add** button
5. Enter your domain name (e.g., `tfdrift-falco.com` or `www.tfdrift-falco.com`)
6. Click **Add**

## Step 2: Configure DNS Records

Vercel will provide you with DNS records that need to be configured at your domain registrar. You'll typically need to add one of the following:

### Option A: Using A Records (for root domain)

If you're setting up a root domain (e.g., `tfdrift-falco.com`):

```
Type: A
Name: @
Value: 76.76.21.21
TTL: 3600
```

### Option B: Using CNAME Record (for subdomain)

If you're setting up a subdomain (e.g., `www.tfdrift-falco.com`):

```
Type: CNAME
Name: www
Value: cname.vercel-dns.com
TTL: 3600
```

### Option C: Using Vercel DNS (Recommended)

For the easiest setup, you can transfer your nameservers to Vercel:

1. In Vercel Dashboard, select **Use Vercel DNS**
2. Update your domain registrar's nameservers to:
   ```
   ns1.vercel-dns.com
   ns2.vercel-dns.com
   ```
3. Wait for DNS propagation (can take up to 48 hours, usually within minutes)

## Step 3: Wait for DNS Propagation

After configuring DNS records:

1. DNS changes can take up to 48 hours to propagate (typically 5-30 minutes)
2. Vercel will automatically issue an SSL certificate once DNS is configured
3. Check your domain status in the Vercel Dashboard

## Step 4: Configure Environment Variables

Update your environment variables to use your custom domain:

1. In Vercel Dashboard, go to **Settings** > **Environment Variables**
2. Update or add:
   ```
   NEXT_PUBLIC_SITE_URL=https://yourdomain.com
   ```
3. Redeploy your application to apply changes

## Step 5: Set Primary Domain (Optional)

If you have multiple domains (e.g., both `tfdrift-falco.com` and `www.tfdrift-falco.com`):

1. In **Domains** settings, click on the domain you want as primary
2. Click **Set as Primary Domain**
3. Non-primary domains will automatically redirect to the primary

## Step 6: Verify Configuration

1. Visit your custom domain in a browser
2. Verify that:
   - The site loads correctly
   - SSL certificate is valid (🔒 icon in browser)
   - All links and navigation work properly
   - Google Analytics is tracking visits (if configured)

## Common DNS Configurations

### Example 1: Root Domain Only

```
# At your DNS provider
Type: A
Name: @
Value: 76.76.21.21

# Result: https://tfdrift-falco.com
```

### Example 2: WWW Subdomain Only

```
# At your DNS provider
Type: CNAME
Name: www
Value: cname.vercel-dns.com

# Result: https://www.tfdrift-falco.com
```

### Example 3: Both Root and WWW (Recommended)

```
# At your DNS provider
Type: A
Name: @
Value: 76.76.21.21

Type: CNAME
Name: www
Value: cname.vercel-dns.com

# Both work: https://tfdrift-falco.com and https://www.tfdrift-falco.com
```

## Troubleshooting

### Domain Not Verifying

- **Check DNS propagation**: Use [DNS Checker](https://dnschecker.org) to verify records
- **Wait longer**: DNS can take up to 48 hours
- **Verify records**: Double-check spelling and values
- **Remove conflicting records**: Delete any existing A/CNAME records for the same hostname

### SSL Certificate Issues

- Vercel automatically issues Let's Encrypt certificates
- Certificates are issued after DNS verification
- If SSL fails, try removing and re-adding the domain

### Redirect Issues

- Ensure only one domain is set as primary
- Clear browser cache and try incognito mode
- Check Vercel deployment logs for errors

## Updating OpenGraph and Metadata

After setting up your custom domain, update SEO metadata in `app/layout.tsx`:

```tsx
export const metadata: Metadata = {
  // ... other metadata
  openGraph: {
    url: "https://yourdomain.com",  // Update this
    // ... other openGraph settings
  },
};
```

Also update `.env.example` and your actual `.env` file:

```bash
NEXT_PUBLIC_SITE_URL=https://yourdomain.com
```

## Additional Resources

- [Vercel Custom Domains Documentation](https://vercel.com/docs/concepts/projects/domains)
- [Vercel DNS Documentation](https://vercel.com/docs/concepts/projects/domains/dns)
- [SSL Certificate Troubleshooting](https://vercel.com/docs/concepts/projects/domains/troubleshooting)

## Support

If you encounter issues:

1. Check [Vercel Status](https://www.vercel-status.com/)
2. Visit [Vercel Community](https://github.com/vercel/vercel/discussions)
3. Contact your domain registrar for DNS-specific questions
4. Open an issue in the [TFDrift-Falco repository](https://github.com/higakikeita/tfdrift-falco/issues)
