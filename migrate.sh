#!/usr/bin/env bash
# =============================================================================
# migrate.sh —— 从 V1 结构迁移到 V2.1（执行 011 → 012 → 013 → 014）
#
# 用法：
#   1) 默认：读取 backend/configs/config.yaml 中的 database 段
#        ./migrate.sh
#   2) 用环境变量覆盖（环境变量优先级最高，其次 config.yaml，再次内置默认值）：
#        DB_HOST=127.0.0.1 DB_PORT=3307 DB_USER=root DB_PASSWORD=secret \
#        DB_NAME=schedule_system_v2 ./migrate.sh
#
# 行为：
#   a) 先对受影响表做 mysqldump 备份，写入 backup/<日期>/ 目录（文件名含日期）。
#      受影响表：users duty_records duty_counters schedule_settings
#               user_permissions_temp applications availability
#      - duty_counters 可能已不存在（已迁移过），自动容错跳过；
#      - 其它表若备份失败（含连接失败），立即 exit 1 中止。
#   b) 备份成功后，按序执行迁移：
#        backend/migrations/011_optimize_schema.sql
#        backend/migrations/012_add_hr_tables.sql
#        backend/migrations/013_add_db_logic.sql
#        backend/migrations/014_fix_temp_perm_reason.sql
#      任一步骤失败立即 exit 1 中止。
# =============================================================================

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/backend/configs/config.yaml"

# ---- 默认值 ----
DEF_HOST="localhost"; DEF_PORT="3306"; DEF_USER="root"; DEF_PASS=""; DEF_NAME="schedule_system_v2"

# ---- 从 config.yaml 的 database 段读取 ----
cfg_val() {
  # $1 = 键名；仅在 database: 段随后的若干行内匹配
  [[ -f "$CONFIG_FILE" ]] || return 0
  grep -A10 "^[[:space:]]*database:" "$CONFIG_FILE" \
    | grep -E "^[[:space:]]*$1:" | head -1 \
    | sed -E "s/^[^:]*:[[:space:]]*\"?([^\"]*)\"?/\1/" | tr -d ' '
}

CFG_HOST="$(cfg_val host)"; CFG_PORT="$(cfg_val port)"
CFG_USER="$(cfg_val user)"; CFG_PASS="$(cfg_val password)"; CFG_NAME="$(cfg_val dbname)"

# ---- 环境变量 > config.yaml > 默认值 ----
resolve() { # env, cfg, default
  if [[ -n "${1:-}" ]]; then echo "$1";
  elif [[ -n "${2:-}" ]]; then echo "$2";
  else echo "$3"; fi
}
DB_HOST="$(resolve "$DB_HOST" "$CFG_HOST" "$DEF_HOST")"
DB_PORT="$(resolve "$DB_PORT" "$CFG_PORT" "$DEF_PORT")"
DB_USER="$(resolve "$DB_USER" "$CFG_USER" "$DEF_USER")"
DB_PASS="$(resolve "$DB_PASS" "$CFG_PASS" "$DEF_PASS")"
DB_NAME="$(resolve "$DB_NAME" "$CFG_NAME" "$DEF_NAME")"

# ---- 拼连接参数 ----
CONN_ARGS=(-h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" --protocol=TCP --default-character-set=utf8mb4)
if [[ -n "$DB_PASS" ]]; then CONN_ARGS+=(-p"$DB_PASS"); fi

echo "========================================================"
echo " 迁移目标数据库: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "========================================================"

MYSQL="mysql"; MYDUMP="mysqldump"
TABLES=(users duty_records duty_counters schedule_settings user_permissions_temp applications availability)
DATE="$(date +%Y%m%d)"
BACKUP_DIR="${SCRIPT_DIR}/backup/${DATE}"

mkdir -p "$BACKUP_DIR" || { echo "[ERROR] 无法创建备份目录 $BACKUP_DIR"; exit 1; }

echo
echo ">>> [1/2] 备份受影响表到 ${BACKUP_DIR}/"
for tbl in "${TABLES[@]}"; do
  # 检查表是否存在
  if ! "$MYSQL" "${CONN_ARGS[@]}" -N -e "SHOW TABLES LIKE '${tbl}';" "$DB_NAME" 2>/dev/null | grep -q "$tbl"; then
    if [[ "$tbl" == "duty_counters" ]]; then
      echo "  - 表 ${tbl} 不存在，跳过（容错：可能已迁移）"
    else
      echo "  - [WARN] 表 ${tbl} 不存在，跳过"
    fi
    continue
  fi
  out="${BACKUP_DIR}/${tbl}_${DATE}.sql"
  echo "  - 备份 ${tbl} -> ${out}"
  if ! "$MYDUMP" "${CONN_ARGS[@]}" "$DB_NAME" "$tbl" > "$out" 2>"${BACKUP_DIR}/.err"; then
    echo "[ERROR] 备份表 ${tbl} 失败，中止迁移："
    cat "${BACKUP_DIR}/.err"
    exit 1
  fi
done
rm -f "${BACKUP_DIR}/.err"
echo "  备份完成。"

echo
echo ">>> [2/2] 执行迁移 011 -> 012 -> 013 -> 014"
for f in 011_optimize_schema.sql 012_add_hr_tables.sql 013_add_db_logic.sql 014_fix_temp_perm_reason.sql; do
  src="${SCRIPT_DIR}/backend/migrations/${f}"
  echo "  - 执行 ${f}"
  if ! "$MYSQL" "${CONN_ARGS[@]}" "$DB_NAME" < "$src" 2>"${BACKUP_DIR}/.err"; then
    echo "[ERROR] 迁移 ${f} 执行失败，中止："
    cat "${BACKUP_DIR}/.err"
    exit 1
  fi
done
rm -f "${BACKUP_DIR}/.err"

echo
echo "========================================================"
echo " 迁移完成。备份位于: ${BACKUP_DIR}/"
echo "========================================================"
