export const timelineData = [
  { date: '03/14', aws: 3, gcp: 1 },
  { date: '03/15', aws: 5, gcp: 2 },
  { date: '03/16', aws: 2, gcp: 3 },
  { date: '03/17', aws: 8, gcp: 1 },
  { date: '03/18', aws: 4, gcp: 4 },
  { date: '03/19', aws: 6, gcp: 2 },
  { date: '03/20', aws: 7, gcp: 3 },
];

export const severityData = [
  { name: 'Critical', value: 5, fill: '#ef4444' },
  { name: 'High', value: 8, fill: '#f97316' },
  { name: 'Medium', value: 12, fill: '#eab308' },
  { name: 'Low', value: 9, fill: '#3b82f6' },
];

export const serviceData = [
  { service: 'Security Group', count: 8 },
  { service: 'IAM Role', count: 6 },
  { service: 'S3 Bucket', count: 5 },
  { service: 'RDS Instance', count: 4 },
  { service: 'Lambda', count: 3 },
  { service: 'KMS Key', count: 3 },
  { service: 'GCE Firewall', count: 3 },
  { service: 'Cloud SQL', count: 2 },
];

export const providerData = [
  { name: 'AWS', value: 24, fill: '#f59e0b' },
  { name: 'GCP', value: 10, fill: '#3b82f6' },
];

export const topUsersData = [
  { user: 'admin-user', events: 12 },
  { user: 'dev@example.com', events: 7 },
  { user: 'ci-bot', events: 5 },
  { user: 'deploy-bot', events: 4 },
  { user: 'ops-user', events: 3 },
];
