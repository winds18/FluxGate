package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RecordTrafficSampleInput struct {
	SampledAt        time.Time
	MetricType       string
	MetricName       string
	RawValueUpload   int64
	RawValueDownload int64
}

type TrafficSample struct {
	ID                 int64  `json:"id"`
	SampledAt          string `json:"sampled_at"`
	MetricType         string `json:"metric_type"`
	MetricName         string `json:"metric_name"`
	UploadBytesDelta   int64  `json:"upload_bytes_delta"`
	DownloadBytesDelta int64  `json:"download_bytes_delta"`
	RawValueUpload     int64  `json:"raw_value_upload"`
	RawValueDownload   int64  `json:"raw_value_download"`
	CreatedAt          string `json:"created_at"`
}

type TokenTrafficSummary struct {
	TokenID            int64  `json:"token_id"`
	UserID             int64  `json:"user_id"`
	GatewayAccountID   int64  `json:"gateway_account_id"`
	AuthUser           string `json:"auth_user"`
	TokenPrefix        string `json:"token_prefix"`
	TokenName          string `json:"token_name"`
	TokenStatus        string `json:"token_status"`
	QuotaBytes         int64  `json:"quota_bytes"`
	TodayUploadBytes   int64  `json:"today_upload_bytes"`
	TodayDownloadBytes int64  `json:"today_download_bytes"`
	TodayTotalBytes    int64  `json:"today_total_bytes"`
	MonthUploadBytes   int64  `json:"month_upload_bytes"`
	MonthDownloadBytes int64  `json:"month_download_bytes"`
	MonthTotalBytes    int64  `json:"month_total_bytes"`
	UsedUploadBytes    int64  `json:"used_upload_bytes"`
	UsedDownloadBytes  int64  `json:"used_download_bytes"`
	UsedTotalBytes     int64  `json:"used_total_bytes"`
	UpdatedAt          string `json:"updated_at"`
}

type TrafficDailySummary struct {
	Day           string `json:"day"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
	TotalBytes    int64  `json:"total_bytes"`
}

func (s *Store) RecordTrafficSample(ctx context.Context, input RecordTrafficSampleInput) (TrafficSample, error) {
	normalized, err := normalizeTrafficSampleInput(input)
	if err != nil {
		return TrafficSample{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TrafficSample{}, err
	}
	defer rollback(tx)

	sample, err := recordTrafficSampleTx(ctx, tx, normalized)
	if err != nil {
		return TrafficSample{}, err
	}
	if err := tx.Commit(); err != nil {
		return TrafficSample{}, err
	}
	return sample, nil
}

func (s *Store) RecordTrafficSamples(ctx context.Context, inputs []RecordTrafficSampleInput) ([]TrafficSample, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	samples := make([]TrafficSample, 0, len(inputs))
	for _, input := range inputs {
		normalized, err := normalizeTrafficSampleInput(input)
		if err != nil {
			return nil, err
		}
		sample, err := recordTrafficSampleTx(ctx, tx, normalized)
		if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return samples, nil
}

func (s *Store) ListTokenTraffic(ctx context.Context) ([]TokenTrafficSummary, error) {
	return s.ListTokenTrafficAt(ctx, time.Now().UTC())
}

func (s *Store) ListTokenTrafficAt(ctx context.Context, now time.Time) ([]TokenTrafficSummary, error) {
	today := now.UTC().Format("2006-01-02")
	monthStart := now.UTC().Format("2006-01") + "-01"
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.user_id, g.id, g.auth_user, t.token_prefix, t.name, t.status,
		       COALESCE(today.upload_bytes, 0), COALESCE(today.download_bytes, 0), COALESCE(today.upload_bytes + today.download_bytes, 0),
		       COALESCE(month.upload_bytes, 0), COALESCE(month.download_bytes, 0), COALESCE(month.upload_bytes + month.download_bytes, 0),
		       t.quota_bytes, t.used_upload_bytes, t.used_download_bytes,
		       t.used_upload_bytes + t.used_download_bytes, t.updated_at
		FROM tokens t
		JOIN gateway_accounts g ON g.token_id = t.id
		LEFT JOIN (
		  SELECT token_id, SUM(upload_bytes) AS upload_bytes, SUM(download_bytes) AS download_bytes
		  FROM traffic_user_daily
		  WHERE day = ?
		  GROUP BY token_id
		) today ON today.token_id = t.id
		LEFT JOIN (
		  SELECT token_id, SUM(upload_bytes) AS upload_bytes, SUM(download_bytes) AS download_bytes
		  FROM traffic_user_daily
		  WHERE day >= ? AND day <= ?
		  GROUP BY token_id
		) month ON month.token_id = t.id
		ORDER BY t.id DESC
	`, today, monthStart, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []TokenTrafficSummary
	for rows.Next() {
		var item TokenTrafficSummary
		if err := rows.Scan(
			&item.TokenID,
			&item.UserID,
			&item.GatewayAccountID,
			&item.AuthUser,
			&item.TokenPrefix,
			&item.TokenName,
			&item.TokenStatus,
			&item.TodayUploadBytes,
			&item.TodayDownloadBytes,
			&item.TodayTotalBytes,
			&item.MonthUploadBytes,
			&item.MonthDownloadBytes,
			&item.MonthTotalBytes,
			&item.QuotaBytes,
			&item.UsedUploadBytes,
			&item.UsedDownloadBytes,
			&item.UsedTotalBytes,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, item)
	}
	return summaries, rows.Err()
}

func (s *Store) ListTrafficDaily(ctx context.Context, days int) ([]TrafficDailySummary, error) {
	return s.ListTrafficDailyAt(ctx, time.Now().UTC(), days)
}

func (s *Store) ListTrafficDailyAt(ctx context.Context, now time.Time, days int) ([]TrafficDailySummary, error) {
	if days <= 0 {
		days = 14
	}
	if days > 90 {
		days = 90
	}
	end := now.UTC()
	start := end.AddDate(0, 0, -days+1)
	rows, err := s.db.QueryContext(ctx, `
		SELECT day, SUM(upload_bytes), SUM(download_bytes)
		FROM traffic_user_daily
		WHERE day >= ? AND day <= ?
		GROUP BY day
	`, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byDay := map[string]TrafficDailySummary{}
	for rows.Next() {
		var item TrafficDailySummary
		if err := rows.Scan(&item.Day, &item.UploadBytes, &item.DownloadBytes); err != nil {
			return nil, err
		}
		item.TotalBytes = item.UploadBytes + item.DownloadBytes
		byDay[item.Day] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	summaries := make([]TrafficDailySummary, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		item := byDay[day]
		item.Day = day
		summaries = append(summaries, item)
	}
	return summaries, nil
}

func normalizeTrafficSampleInput(input RecordTrafficSampleInput) (RecordTrafficSampleInput, error) {
	input.MetricType = strings.ToLower(strings.TrimSpace(input.MetricType))
	input.MetricName = strings.TrimSpace(input.MetricName)
	if input.SampledAt.IsZero() {
		input.SampledAt = time.Now().UTC()
	} else {
		input.SampledAt = input.SampledAt.UTC()
	}
	if input.MetricType != "user" && input.MetricType != "inbound" && input.MetricType != "outbound" {
		return RecordTrafficSampleInput{}, fmt.Errorf("metric_type must be user, inbound or outbound")
	}
	if input.MetricName == "" {
		return RecordTrafficSampleInput{}, fmt.Errorf("metric_name is required")
	}
	if input.RawValueUpload < 0 || input.RawValueDownload < 0 {
		return RecordTrafficSampleInput{}, fmt.Errorf("raw traffic values must be non-negative")
	}
	return input, nil
}

func recordTrafficSampleTx(ctx context.Context, tx *sql.Tx, input RecordTrafficSampleInput) (TrafficSample, error) {
	uploadDelta, downloadDelta, err := trafficDelta(ctx, tx, input)
	if err != nil {
		return TrafficSample{}, err
	}
	sampledAt := input.SampledAt.Format(time.RFC3339)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO traffic_samples(sampled_at, metric_type, metric_name, upload_bytes_delta, download_bytes_delta, raw_value_upload, raw_value_download)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sampledAt, input.MetricType, input.MetricName, uploadDelta, downloadDelta, input.RawValueUpload, input.RawValueDownload)
	if err != nil {
		return TrafficSample{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return TrafficSample{}, err
	}
	if input.MetricType == "user" && (uploadDelta > 0 || downloadDelta > 0) {
		if err := applyUserTrafficDelta(ctx, tx, input.MetricName, input.SampledAt, uploadDelta, downloadDelta); err != nil {
			return TrafficSample{}, err
		}
	}
	if input.MetricType == "outbound" && (uploadDelta > 0 || downloadDelta > 0) {
		if err := applyOutboundTrafficDelta(ctx, tx, input.MetricName, input.SampledAt, uploadDelta, downloadDelta); err != nil {
			return TrafficSample{}, err
		}
	}
	return getTrafficSampleTx(ctx, tx, id)
}

func trafficDelta(ctx context.Context, tx *sql.Tx, input RecordTrafficSampleInput) (int64, int64, error) {
	var previousUpload, previousDownload int64
	err := tx.QueryRowContext(ctx, `
		SELECT raw_value_upload, raw_value_download
		FROM traffic_samples
		WHERE metric_type = ? AND metric_name = ?
		ORDER BY sampled_at DESC, id DESC
		LIMIT 1
	`, input.MetricType, input.MetricName).Scan(&previousUpload, &previousDownload)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return counterDelta(input.RawValueUpload, previousUpload), counterDelta(input.RawValueDownload, previousDownload), nil
}

func counterDelta(current, previous int64) int64 {
	if current >= previous {
		return current - previous
	}
	return current
}

func applyUserTrafficDelta(ctx context.Context, tx *sql.Tx, authUser string, sampledAt time.Time, uploadDelta, downloadDelta int64) error {
	var tokenID, userID, gatewayAccountID int64
	err := tx.QueryRowContext(ctx, `
		SELECT t.id, t.user_id, g.id
		FROM gateway_accounts g
		JOIN tokens t ON t.id = g.token_id
		WHERE g.auth_user = ?
	`, authUser).Scan(&tokenID, &userID, &gatewayAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	sampledAtText := sampledAt.Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, `
		UPDATE tokens
		SET used_upload_bytes = used_upload_bytes + ?,
		    used_download_bytes = used_download_bytes + ?,
		    last_used_at = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, uploadDelta, downloadDelta, sampledAtText, tokenID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tokens
		SET status = 'over_quota', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND status = 'active'
		  AND quota_bytes > 0
		  AND used_upload_bytes + used_download_bytes >= quota_bytes
	`, tokenID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE gateway_accounts
		SET status = 'over_quota', updated_at = CURRENT_TIMESTAMP
		WHERE token_id = ?
		  AND status = 'active'
		  AND EXISTS (
		    SELECT 1 FROM tokens
		    WHERE tokens.id = gateway_accounts.token_id
		      AND tokens.status = 'over_quota'
		  )
	`, tokenID); err != nil {
		return err
	}

	hour := sampledAt.Truncate(time.Hour).Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO traffic_user_hourly(hour, user_id, token_id, gateway_account_id, upload_bytes, download_bytes)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(hour, token_id) DO UPDATE SET
		  upload_bytes = upload_bytes + excluded.upload_bytes,
		  download_bytes = download_bytes + excluded.download_bytes,
		  updated_at = CURRENT_TIMESTAMP
	`, hour, userID, tokenID, gatewayAccountID, uploadDelta, downloadDelta); err != nil {
		return err
	}

	day := sampledAt.Format("2006-01-02")
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO traffic_user_daily(day, user_id, token_id, gateway_account_id, upload_bytes, download_bytes)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(day, token_id) DO UPDATE SET
		  upload_bytes = upload_bytes + excluded.upload_bytes,
		  download_bytes = download_bytes + excluded.download_bytes,
		  updated_at = CURRENT_TIMESTAMP
	`, day, userID, tokenID, gatewayAccountID, uploadDelta, downloadDelta); err != nil {
		return err
	}
	return nil
}

func applyOutboundTrafficDelta(ctx context.Context, tx *sql.Tx, outboundTag string, sampledAt time.Time, uploadDelta, downloadDelta int64) error {
	var upstreamNodeID any
	if id, ok := upstreamNodeIDFromOutboundTag(outboundTag); ok {
		var found int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM upstream_nodes WHERE id = ?", id).Scan(&found)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == nil {
			upstreamNodeID = found
		}
	}
	day := sampledAt.Format("2006-01-02")
	_, err := tx.ExecContext(ctx, `
		INSERT INTO traffic_outbound_daily(day, upstream_node_id, outbound_tag, upload_bytes, download_bytes)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(day, outbound_tag) DO UPDATE SET
		  upload_bytes = upload_bytes + excluded.upload_bytes,
		  download_bytes = download_bytes + excluded.download_bytes,
		  updated_at = CURRENT_TIMESTAMP
	`, day, upstreamNodeID, outboundTag, uploadDelta, downloadDelta)
	return err
}

func upstreamNodeIDFromOutboundTag(tag string) (int64, bool) {
	raw := strings.TrimPrefix(strings.TrimSpace(tag), "up_")
	if raw == tag || raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
}

func getTrafficSampleTx(ctx context.Context, tx *sql.Tx, id int64) (TrafficSample, error) {
	var sample TrafficSample
	err := tx.QueryRowContext(ctx, `
		SELECT id, sampled_at, metric_type, metric_name, upload_bytes_delta, download_bytes_delta, raw_value_upload, raw_value_download, created_at
		FROM traffic_samples
		WHERE id = ?
	`, id).Scan(
		&sample.ID,
		&sample.SampledAt,
		&sample.MetricType,
		&sample.MetricName,
		&sample.UploadBytesDelta,
		&sample.DownloadBytesDelta,
		&sample.RawValueUpload,
		&sample.RawValueDownload,
		&sample.CreatedAt,
	)
	return sample, err
}
