<script setup lang="ts">
import PackInfoForm from "@/components/pack/PackInfoForm.vue";
import type {NewPackRequest} from "@/interfaces/requests.ts";
import {newPack} from "@/services/packs.service.ts";
import {sleep} from "@/services/utils.ts";

const creating = ref(false)
const error = ref(false)

const data = ref({
  slug: '',
  name: '',
  packVersion: "0.1.0",
  description: '',
  minecraftVersion: "",
  loader: {
    name: undefined,
    version: '',
  },
  acceptableVersions: [],
})

const router = useRouter()

const buildRequest: () => NewPackRequest = () => {
  const form = data.value

  const nonVersion = ["Latest", "LatestSnapshot"]

  return {
    slug: form.slug,
    name: form.name,
    version: form.packVersion,
    description: form.description,
    minecraft: {
      version: nonVersion.includes(form.minecraftVersion || "") ? "" : form.minecraftVersion || "",
      latest: form.minecraftVersion === "Latest",
      snapshot: form.minecraftVersion === "Latest Snapshot",
    },
    loader: {
      name: (form.loader.name || "").toLowerCase(),
      version: form.loader.version === "Latest" ? "" : form.loader.version,
      latest: form.loader.version === "Latest",
    },
    acceptableVersions: form.acceptableVersions,
  }
}

const submitForm = async () => {
  error.value = false
  creating.value = true
  const request = buildRequest()

  try {
    await newPack(request)

    await sleep(1500)
    await router.push({path: `/packs/${request.slug}`})

  } catch (e) {
    error.value = true
    console.error(e)
    return
  } finally {
    creating.value = false
  }
}

const cancelForm = async () => {
  await router.push({path: "/packs"})
}

</script>

<template>
  <div
    class="ma-6"
  >
    <v-alert
      v-if="error"
      class="mb-6"
      text="Error creating new pack..."
      type="error"
      icon="mdi-alert"
      closable
    />
    <PackInfoForm
      v-model:data="data"
      v-model:loading="creating"
      title="New Pack"
      @submit-data="submitForm"
      @cancel-op="cancelForm"
    />
  </div>
</template>
