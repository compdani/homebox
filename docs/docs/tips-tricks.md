# Tips and Tricks

## Custom Fields

Custom fields are a great way to add any extra information to your item. The following types are supported:

- [x] Text
- [ ] Integer (Future)
- [ ] Boolean (Future)
- [ ] Timestamp (Future)

Custom fields are appended to the main details section of your item.

!!! tip
    Homebox Custom Fields also have special support for URLs. Provide a URL (`https://google.com`) and it will be automatically converted to a clickable link in the UI. Optionally, you can also use Markdown syntax to add a custom text to the button. `[Google](https://google.com)`

## Managing Asset IDs

Homebox provides the option to auto-set asset IDs, this is the default behavior. These can be used for tracking assets with printable tags or labels. You can disable this behavior via a command line flag or ENV variable. See [configuration](../quick-start#env-variables-configuration) for more details.

Example ID: `000-001`

Asset IDs are partially managed by Homebox, but have a flexible implementation to allow for unique use cases. IDs are non-unique at the database level, so there is nothing stopping a user from manually setting duplicate IDs for various items. There are two recommended approaches to manage Asset IDs:

### 1. Auto Incrementing IDs

This is the default behavior likely to experience the most consistency. Whenever creating or importing an item, that item receives the next available ID. This is recommended for most users.

### 2. Auto Incrementing IDs with Reset

In some cases, you may want to skip some items such as consumables, or items that are loosely tracked. In this case, we recommend that you leave auto-incrementing IDs enabled _however_ when you create a new item that you want to skip, you can go to that item and reset the ID to 0. This will remove it from the auto-incrementing sequence, and the next item will receive the next available ID.

!!! tip
    If you're migrating from an older version, there is an action on the user's profile page to assign IDs to all items. This will assign the next available ID to all items in order of their creation. You should __only do this once__ during the migration process. You should be especially cautious with this if you're using the reset feature described in [option number 2](#2-auto-incrementing-ids-with-reset)

## QR Codes

:octicons-tag-24: 0.7.0

Homebox generates QR codes for **products**, **locations**, and **unique items** (items not linked to a product). Product-linked placements use the product QR code.

### Printable labels (PNG)

Download ready-to-print PNG labels from entity detail pages:

| Entity | Label size (landscape) | QR encodes |
|--------|------------------------|------------|
| Location | 3.15" × 1.97" | `/location/{id}` |
| Product | 1.57" × 1.18" | `/product/{id}` |
| Unique item | 1.57" × 1.18" | `/item/{id}` |

API endpoints (authenticated): `/api/v1/locations/{id}/label.png`, `/api/v1/products/{id}/label.png`, `/api/v1/items/{id}/label.png` (unique items only).

### In-app scanning

Use **Scan** in the sidebar for three workflows:

1. **Product then location** — scan a product (or unique item), then a location, then enter quantity.
2. **Location then batch** — scan a location, then scan multiple products/items; quantity is prompted after each scan.
3. **Remove product from location** — scan a product, then the location to remove from, then enter how many to remove. If you remove all units, the placement is deleted.

Manual adds from **Create → Product to Location** always require a quantity.

### Raw QR API

The API endpoint `GET /api/v1/qrcode?data=...` generates a QR image for any URL-encoded payload.

:octicons-tag-24: v0.8.0

The tools page still includes the asset-ID label generator for pre-printed tags (`/a/{assetId}` URLs). That flow is separate from product/location labels.

[Demo](https://homebox.fly.dev/reports/label-generator)

## Scheduled Maintenance Notifications

:octicons-tag-24: v0.9.0

Homebox uses [shoutrrr](https://containrrr.dev/shoutrrr/0.7/) to send notifications. This allows you to send notifications to a variety of services. On your profile page, you can add notification URLs to your profile which will be used to send notifications when a maintenance event is scheduled.

**Notifications are sent on the day the maintenance is scheduled at or around 8am.**

As of `v0.9.0` we have limited support for complex scheduling of maintenance events. If you have requests for extended functionality, please open an issue on GitHub or reach out on Discord. We're still gauging the demand for this feature.


## Custom Currencies

:octicons-tag-24: v0.11.0

Homebox allows you to add additional currencies to your instance by specify a JSON file containing the currencies you want to add.

**Environment Variable:** `HBOX_OPTIONS_CURRENCY_CONFIG`

### Example

```json
[
  {
    "code": "AED",
    "local": "United Arab Emirates",
    "symbol": "د.إ",
    "name": "United Arab Emirates Dirham"
  },
]
```
