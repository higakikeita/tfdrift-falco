# GCP BigQuery

> **Service:** BigQuery
> **Events Monitored:** 6
> **Resources:** `google_bigquery_dataset`, `google_bigquery_table`
> **Status:** ✅ Production Ready

## Monitored Events

### Datasets (3 events)
- `google.cloud.bigquery.v2.DatasetService.InsertDataset` - Dataset creation
- `google.cloud.bigquery.v2.DatasetService.DeleteDataset` - Dataset deletion
- `google.cloud.bigquery.v2.DatasetService.PatchDataset` - Dataset update

### Tables (3 events)
- `google.cloud.bigquery.v2.TableService.InsertTable` - Table creation
- `google.cloud.bigquery.v2.TableService.DeleteTable` - Table deletion
- `google.cloud.bigquery.v2.TableService.PatchTable` - Table update
