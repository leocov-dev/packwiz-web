<script setup lang="ts">
import PackInfoForm from "@/components/pack/PackInfoForm.vue";
import type {EditPackRequest} from "@/interfaces/requests.ts";
import {sleep} from "@/services/utils.ts";
import {type Pack} from "@/interfaces/pack.ts";
import {editPack} from "@/services/packs.service.ts";
import {LATEST_SENTINEL, LATEST_SNAPSHOT_SENTINEL} from "@/components/forms/MinecraftVersion.vue";

const {pack} = defineProps<{ pack: Pack }>()

const editing = ref(false)
const error = ref(false)

const data = ref({
  id: pack.id,
  slug: pack.slug,
  name: pack.name,
  packVersion: pack.version,
  description: pack.description,
  minecraftVersion: pack.mcVersion,
  loader: {
    name: pack.loader,
    version: pack.loaderVersion,
  },
  acceptableVersions: pack.acceptableGameVersions || [],
})

const router = useRouter()

const buildRequest: () => EditPackRequest = () => {
  const form = data.value

  const isLatest = form.minecraftVersion === LATEST_SENTINEL
  const isSnapshot = form.minecraftVersion === LATEST_SNAPSHOT_SENTINEL

  return {
    name: form.name,
    version: form.packVersion,
    description: form.description,
    minecraft: {
      version: (isLatest || isSnapshot) ? "" : form.minecraftVersion || "",
      latest: isLatest,
      snapshot: isSnapshot,
    },
    loader: {
      name: (form.loader.name || "").toLowerCase(),
      version: form.loader.version || "",
      latest: false,
    },
    acceptableVersions: form.acceptableVersions,
  }
}

const submitForm = async () => {
  error.value = false
  editing.value = true
  const request = buildRequest()

  try {
    await editPack(pack.id, request)

    await sleep(1500)
    await router.push({path: `/packs/${pack.id}`})

  } catch (e) {
    error.value = true
    console.error(e)
    return
  } finally {
    editing.value = false
  }
}

const cancelForm = async () => {
  await router.push({path: `/packs/${pack.id}`})
}

</script>

<template>
  <div
    class="ma-6"
  >
    <v-alert
      v-if="error"
      class="mb-6"
      text="Error editing pack..."
      type="error"
      icon="mdi-alert"
      closable
    />
    <PackInfoForm
      v-model:data="data"
      v-model:loading="editing"
      title="Edit Pack"
      accept-text="Save"
      slug-locked
      @submit-data="submitForm"
      @cancel-op="cancelForm"
    />
  </div>
</template>
