# PMOVES.AI Tier Environment: Data
# For Data services (Qdrant, Neo4j, Meilisearch, MinIO, etc.)
# Source after env.shared: source env.shared && source env.tier-data.sh

# ============================================================================
# Data Tier Configuration
# ============================================================================

export TIER=data

# Connection limits
export MAX_CONNECTIONS=${MAX_CONNECTIONS:-100}
export MAX_IDLE_CONNECTIONS=${MAX_IDLE_CONNECTIONS:-10}
export CONNECTION_TIMEOUT_MS=${CONNECTION_TIMEOUT_MS:-10000}
export IDLE_TIMEOUT_MS=${IDLE_TIMEOUT_MS:-300000}

# Qdrant (Vector Database)
export QDRANT_COLLECTION=${QDRANT_COLLECTION:-pmoves_chunks}
export QDRANT_VECTOR_SIZE=${QDRANT_VECTOR_SIZE:-384}
export QDRANT_DISTANCE=${QDRANT_DISTANCE:-Cosine}
export QDRANT_REPLICATION_FACTOR=${QDRANT_REPLICATION_FACTOR:-1}

# Neo4j (Graph Database)
export NEO4J_DATABASE=${NEO4J_DATABASE:-neo4j}
export NEO4J_MAX_TRANSACTION_RUNTIME_MS=${NEO4J_MAX_TRANSACTION_RUNTIME_MS:-60000}
export NEO4J_CONNECTION_POOL_SIZE=${NEO4J_CONNECTION_POOL_SIZE:-50}

# Meilisearch (Full-Text Search)
export MEILISEARCH_INDEX=${MEILISEARCH_INDEX:-pmoves_docs}
export MEILISEARCH_TYPO_TOLERANCE=${MEILISEARCH_TYPO_TOLERANCE:-true}
export MEILISEARCH_MIN_WORD_SIZE_FOR_TYPOS=${MEILISEARCH_MIN_WORD_SIZE_FOR_TYPOS:-4}

# MinIO (Object Storage)
export MINIO_BUCKET_ASSETS=${MINIO_BUCKET_ASSETS:-assets}
export MINIO_BUCKET_OUTPUTS=${MINIO_BUCKET_OUTPUTS:-outputs}
export MINIO_PRESIGN_TTL=${MINIO_PRESIGN_TTL:-3600}  # 1 hour
export MINIO_MAX_PART_SIZE=${MINIO_MAX_PART_SIZE:-5242880}  # 5MB

# Backup configuration
export BACKUP_ENABLED=${BACKUP_ENABLED:-true}
export BACKUP_INTERVAL=${BACKUP_INTERVAL:-86400}  # 24 hours
export BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-7}
