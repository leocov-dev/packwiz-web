DELETE
FROM permissions
WHERE name = 'pack.webhook.manage';

DROP TABLE pack_webhooks;
