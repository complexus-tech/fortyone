# Redis Streams Migration

This document outlines the completed migration from Redis Pub/Sub to Redis Streams for event processing.

## Background

Our application was initially using Redis Pub/Sub for event distribution. While this worked well for a single instance, it led to duplicate event processing when running multiple instances since each subscriber receives all messages.

Redis Streams provides a more robust solution for event processing with consumer groups, which ensures each event is processed by only one consumer.

## Implementation Status

The migration was designed to be non-disruptive with a phased approach:

1. **Phase 1: Dual Publishing and Consumption** ✅

   - Events are published to both Pub/Sub channels and a Redis Stream
   - Both the legacy Pub/Sub consumer and the new Streams consumer run concurrently
   - The legacy consumer can be disabled via configuration (`APP_EVENTS_USE_LEGACY_CONSUMER=false`)

2. **Phase 2: Transition to Streams Only** ✅

   - After confirming stability, disabled legacy consumer on all environments
   - Removed Pub/Sub publishing from the Publisher

3. **Phase 3: Code Cleanup** ✅
   - Removed legacy consumer code entirely
   - Simplified publisher to only use Streams

## MIGRATION COMPLETE ✅

The application now exclusively uses Redis Streams for event processing:

- Events are published to a single Redis Stream (`events-stream`)
- Events are consumed by instances in a consumer group (`events-processors`)
- Each consumer instance has a unique identifier
- Message acknowledgment ensures exactly-once processing
- Pending message handling recovers from failures

## Monitoring

Monitor the following metrics to ensure the system is working properly:

1. Redis memory usage
2. Event processing metrics
3. Event processing latency

## Commands for Monitoring

```bash
# View stream information
redis-cli xinfo stream events-stream

# View consumer groups
redis-cli xinfo groups events-stream

# View pending messages
redis-cli xpending events-stream events-processors

# View stream content
redis-cli xrange events-stream - + COUNT 10
```

## Benefits of Redis Streams

1. **Exactly-once processing**: Events are processed exactly once, eliminating duplicates
2. **Persistence**: Messages are stored until acknowledged
3. **Fault tolerance**: Pending message handling ensures no events are lost
4. **Scalability**: Multiple instances can process events efficiently

## Configuration

- `APP_EVENTS_USE_LEGACY_CONSUMER`: Controls whether the legacy Pub/Sub consumer runs
  - Default: `true` (for backward compatibility)
  - Set to `false` to disable Pub/Sub consumption

## Monitoring the Transition

During the transition, monitor the following:

1. Redis memory usage
2. Event processing metrics
3. Duplicate event processing
4. Event processing latency

## Rollback Plan

If issues are encountered:

1. Ensure `APP_EVENTS_USE_LEGACY_CONSUMER=true` to fallback to the original behavior
2. If severe issues with streams, modify the publisher to stop publishing to streams

## Testing Recommendation

1. Test with a single instance first
2. Then test with multiple instances to verify events are processed exactly once
3. Monitor for any missed events

## Timeline

- Phase 1: Implement dual approach (complete)
- Phase 2: Transition period (2-4 weeks)
- Phase 3: Cleanup (after successful transition)

## Technical Reference

- Redis Stream Key: `events-stream`
- Consumer Group: `events-processors`
- Each consumer instance gets a unique UUID identifier
