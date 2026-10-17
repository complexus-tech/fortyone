//go:build integration

package mayarepository

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSharedMayaEntitlementTracksTrialPaidRevocationAndDeletion(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace := uuid.New()
	_, err := postgres.Pool.Exec(ctx, `INSERT INTO workspaces(workspace_id,name,slug,trial_ends_on) VALUES($1,'Entitlement test',$2,clock_timestamp()+interval '1 day')`, workspace, uuid.NewString())
	require.NoError(t, err)
	check := func(want bool) {
		t.Helper()
		var allowed bool
		require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT public.workspace_can_use_maya($1)`, workspace).Scan(&allowed))
		require.Equal(t, want, allowed)
	}
	check(true)
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspaces SET trial_ends_on=clock_timestamp()-interval '1 day' WHERE workspace_id=$1`, workspace)
	require.NoError(t, err)
	check(false)
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO workspace_subscriptions(workspace_id,stripe_customer_id,stripe_subscription_id,subscription_tier,subscription_status) VALUES($1,'test','test','pro','active')`, workspace)
	require.NoError(t, err)
	check(true)
	for _, status := range []string{"trialing", "past_due"} {
		_, err = postgres.Pool.Exec(ctx, `UPDATE workspace_subscriptions SET subscription_status=$2 WHERE workspace_id=$1`, workspace, status)
		require.NoError(t, err)
		check(true)
	}
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspace_subscriptions SET subscription_status='canceled' WHERE workspace_id=$1`, workspace)
	require.NoError(t, err)
	check(false)
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspace_subscriptions SET subscription_status='active',subscription_tier='free' WHERE workspace_id=$1`, workspace)
	require.NoError(t, err)
	check(false)
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspaces SET trial_ends_on=clock_timestamp()+interval '1 day',deleted_at=clock_timestamp() WHERE workspace_id=$1`, workspace)
	require.NoError(t, err)
	check(false)
}
