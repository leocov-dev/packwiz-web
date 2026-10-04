<script setup lang="ts">
import PackInfoForm from "@/components/pack/PackInfoForm.vue";
import type {EditPackRequest} from "@/interfaces/requests.ts";
import {sleep} from "@/services/utils.ts";
import {type Pack} from "@/interfaces/pack.ts";
import {editPack} from "@/services/packs.service.ts";
import {LATEST_SENTINEL, LATEST_SNAPSHOT_SENTINEL} from "@/components/forms/MinecraftVersion.vue";
import {checkTargetChange} from "@/lib/target-guard.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import ConsumerImpactDialog from "@/components/pack/ConsumerImpactDialog.vue";

const {pack} = defineProps<{ pack: Pack }>()

const editing = ref(false)
const error = ref<string | null>(null)

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

// published packs can't switch loader or go to an older Minecraft version,
// and a Minecraft change is confirmed against who uses the pack
const targetCheck = computed(() =>
  checkTargetChange(pack, data.value.minecraftVersion || "", data.value.loader.name || ""))
const showConfirm = ref(false)

const requestSubmit = () => {
  if (targetCheck.value.blocked) return
  if (targetCheck.value.confirm) {
    showConfirm.value = true
    return
  }
  void submitForm()
}

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
  error.value = null
  editing.value = true
  const request = buildRequest()

  try {
    await editPack(pack.id, request)

    await sleep(1500)
    await router.push({path: `/packs/${pack.id}`})

  } catch (e) {
    error.value = apiErrorMessage(e, "Error editing pack...")
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
      :text="error"
      type="error"
      icon="mdi-alert"
      closable
    />
    <v-alert
      v-if="targetCheck.blocked"
      class="mb-6"
      :text="targetCheck.blocked"
      type="error"
      variant="tonal"
      icon="mdi-lock"
    />
    <ConsumerImpactDialog
      v-model="showConfirm"
      :pack-id="pack.id"
      title="Change the Minecraft version?"
      accept-text="Save"
      danger
      @accepted="submitForm"
    >
      <p>
        This published pack moves from Minecraft <strong>{{ pack.mcVersion }}</strong> to
        <strong>{{ data.minecraftVersion }}</strong>. On their next launch, players' instances
        will be asked to switch. Their worlds are upgraded and can't be opened in the old version
        again, and the server has to move to the same version.
      </p>
    </ConsumerImpactDialog>
    <PackInfoForm
      v-model:data="data"
      v-model:loading="editing"
      title="Edit Pack"
      accept-text="Save"
      slug-locked
      @submit-data="requestSubmit"
      @cancel-op="cancelForm"
    />
  </div>
</template>
