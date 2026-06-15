import { COLLECTIONS, configurePb, getPb } from "~~/lib/pocketbase/client";

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();
  if (config.public.pbUrl) {
    configurePb(config.public.pbUrl);
  }
  const pb = getPb();
  pb.authStore.onChange(() => {
    // keep reactive auth in sync
  });
  return {
    provide: {
      pb,
    },
  };
});
