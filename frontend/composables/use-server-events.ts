import { COLLECTIONS, getPb } from "~~/lib/pocketbase/client";

export enum ServerEvent {
  LocationMutation = "location.mutation",
  ItemMutation = "item.mutation",
  LabelMutation = "label.mutation",
}

export type EventMessage = {
  event: ServerEvent;
};

const listeners = new Map<ServerEvent, (() => void)[]>();
let subscribed = false;

function ensureSubscriptions(onmessage: (m: EventMessage) => void) {
  if (subscribed) {
    return;
  }
  subscribed = true;
  const pb = getPb();

  const throttled = new Map<ServerEvent, any>();
  throttled.set(ServerEvent.LocationMutation, useThrottleFn(onmessage, 1000));
  throttled.set(ServerEvent.ItemMutation, useThrottleFn(onmessage, 1000));
  throttled.set(ServerEvent.LabelMutation, useThrottleFn(onmessage, 1000));

  const notify = (event: ServerEvent) => {
    const fn = throttled.get(event);
    fn?.({ event });
    listeners.get(event)?.forEach(c => c());
  };

  const subscribe = (collection: string, event: ServerEvent) => {
    pb.collection(collection)
      .subscribe("*", () => notify(event))
      .catch(err => {
        console.warn(`realtime subscription failed for ${collection}`, err);
      });
  };

  subscribe(COLLECTIONS.items, ServerEvent.ItemMutation);
  subscribe(COLLECTIONS.locations, ServerEvent.LocationMutation);
  subscribe(COLLECTIONS.labels, ServerEvent.LabelMutation);
}

export function onServerEvent(event: ServerEvent, callback: () => void) {
  ensureSubscriptions(e => {
    console.debug("received event", e);
    listeners.get(e.event)?.forEach(c => c());
  });

  onMounted(() => {
    if (!listeners.has(event)) {
      listeners.set(event, []);
    }
    listeners.get(event)?.push(callback);
  });

  onUnmounted(() => {
    const got = listeners.get(event);
    if (got) {
      listeners.set(
        event,
        got.filter(c => c !== callback)
      );
    }
    if (listeners.get(event)?.length === 0) {
      listeners.delete(event);
    }
  });
}
