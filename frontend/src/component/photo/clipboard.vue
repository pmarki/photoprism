<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard p-photo-clipboard`"
        :end="!rtl"
        :start="rtl"
        :attach="true"
        location="top"
        transition="slide-y-reverse-transition"
        offset="12"
      >
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            icon
            size="52"
            color="highlight"
            variant="elevated"
            density="comfortable"
            class="action-menu opacity-95 ma-5"
          >
            <span class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn
          v-if="canShare && canServiceUpload && context !== 'archive' && context !== 'hidden' && context !== 'review'"
          key="action-share"
          :title="$gettext('Share')"
          prepend-icon="mdi-share"
          color="share"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-share"
          @click.stop="dialog.share = true"
          >{{ $gettext("Share") }}</v-btn
        >
        <v-btn
          v-if="canManage && context === 'review'"
          key="action-approve"
          :title="$gettext('Approve')"
          prepen-icon="mdi-check-bold"
          color="share"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-approve"
          @click.stop="batchApprove"
          >{{ $gettext("Approve") }}</v-btn
        >
        <v-btn
          v-if="canArchive && !album && context === 'archive' && context !== 'hidden'"
          key="action-restore"
          :title="$gettext('Restore')"
          prepend-icon="mdi-archive-arrow-up"
          color="share"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-restore"
          @click.stop="batchRestore"
          >{{ $gettext("Restore") }}</v-btn
        >
<!--        <v-btn-->
<!--          v-if="canEdit"-->
<!--          key="action-edit"-->
<!--          :title="$gettext('Edit')"-->
<!--          prepend-icon="mdi-pencil"-->
<!--          color="edit"-->
<!--          variant="elevated"-->
<!--          density="comfortable"-->
<!--          :disabled="selection.length === 0 || busy"-->
<!--          class="action-edit"-->
<!--          @click.stop="edit"-->
<!--          >{{ $gettext("Edit") }}</v-btn-->
<!--        >-->
        <v-btn
          v-if="canTogglePrivate && context !== 'archive' && context !== 'hidden'"
          key="action-private"
          :title="$gettext('Change private flag')"
          icon="mdi-lock"
          color="private"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-private"
          @click.stop="batchPrivate"
        ></v-btn>
        <v-btn
          v-if="canDownload && !navigatorCanShare && context !== 'archive'"
          key="action-download"
          :title="$gettext('Download')"
          prepend-icon="mdi-download"
          color="download"
          variant="elevated"
          density="comfortable"
          :disabled="busy"
          class="action-download"
          @click.stop="download()"
          >{{ $gettext("Download") }}</v-btn
        >
        <v-btn
          v-if="canDownload && context !== 'archive'"
          key="action-webshare"
          :title="$gettext('Share')"
          prepend-icon="mdi-share-variant"
          color="download"
          variant="elevated"
          density="comfortable"
          :disabled="busy"
          class="action-webshare"
          @click.stop="webShare"
          >{{ $gettext("Share") }}</v-btn
        >
        <v-btn
          v-if="canEditAlbum && context !== 'archive' && context !== 'hidden'"
          key="action-album"
          :title="$gettext('Add to album')"
          prepend-icon="mdi-bookmark"
          color="album"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-album"
          @click.stop="dialog.album = true"
          >{{ $gettext("Add to album") }}</v-btn
        >
        <v-btn
          v-if="canArchive && context !== 'archive' && context !== 'hidden'"
          key="action-archive"
          :title="$gettext('Archive')"
          icon="mdi-archive"
          color="remove"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-archive"
          @click.stop="archivePhotos"
        ></v-btn>
        <v-btn
          v-if="canEditAlbum && isAlbum"
          key="action-remove"
          :title="$gettext('Remove from Album')"
          prepend-icon="mdi-eject"
          color="remove"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-remove"
          @click.stop="removeFromAlbum"
        >{{ $gettext("Remove from Album") }}</v-btn
        >
        <v-btn
          v-if="canDelete"
          key="action-move"
          :title="$gettext('Move to folder')"
          prepend-icon="mdi-eject"
          color="remove"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-move"
          @click.stop="dialog.move = true"
          >{{ $gettext("Move to folder") }}</v-btn
        >
        <v-btn
          v-if="canDelete"
          key="action-delete"
          :title="$gettext('Delete')"
          prepend-icon="mdi-delete"
          color="remove"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0 || busy"
          class="action-delete"
          @click.stop="deletePhotos"
          >{{ $gettext("Delete") }}</v-btn
        >
        <v-btn
          key="action-close"
          prepend-icon="mdi-close"
          color="grey-darken-2"
          variant="elevated"
          density="comfortable"
          class="action-clear"
          @click.stop="clearClipboard()"
          >{{ $gettext("Unselect") }}</v-btn
        >
      </v-speed-dial>
    </div>
    <p-photo-archive-dialog
      :visible="dialog.archive"
      @close="dialog.archive = false"
      @confirm="batchArchive"
    ></p-photo-archive-dialog>
    <p-confirm-dialog
      :visible="dialog.delete"
      :text="$gettext(`Are you sure you want to permanently delete these pictures?`)"
      :action="$gettext('Yes')"
      icon="mdi-delete-outline"
      @close="dialog.delete = false"
      @confirm="batchDelete"
    ></p-confirm-dialog>
    <p-photo-album-dialog
      :visible="dialog.album"
      @close="dialog.album = false"
      @confirm="addToAlbum"
    ></p-photo-album-dialog>
    <p-service-upload
      :visible="dialog.share"
      :items="{ photos: selection }"
      :model="album"
      @close="dialog.share = false"
      @confirm="onShared"
    ></p-service-upload>
    <p-webshare-dialog
      :show="dialog.webshare"
      :items="{ photos: selection }"
      @completed="
        busy = false;
        dialog.webshare = false;
      "
      @failed="onWebshareFailed"
    ></p-webshare-dialog>
    <p-move-dialog
      :visible="dialog.move"
      :selection="selection"
      @close="dialog.move = false"
      @confirm="dialog.move = false"
    ></p-move-dialog>
  </div>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import download from "common/download";
import Photo from "model/photo";
import { File as PFile } from "model/file";
import { canUseWebshareApi } from "common/can";

import PConfirmDialog from "component/confirm/dialog.vue";
import PPhotoAlbumDialog from "component/photo/album/dialog.vue";
import Api from "../../common/api";
import { $config } from "../../app/session";
import PMoveDialog from "./move/dialog.vue";

export default {
  name: "PPhotoClipboard",
  components: {
    PMoveDialog,
    PConfirmDialog,
    PPhotoAlbumDialog,
  },
  props: {
    context: {
      type: String,
      default: "photos",
    },
    refresh: {
      type: Function,
      default: () => {},
    },
    album: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    const features = this.$config.getSettings().features;

    return {
      selection: this.$clipboard.selection,
      canTogglePrivate: this.$config.allow("photos", "manage") && features.private,
      canArchive: this.$config.allow("photos", "delete") && features.archive,
      canDelete: this.$config.allow("photos", "delete") && features.delete,
      canDownload: this.$config.allow("photos", "download") && features.download,
      canShare: this.$config.allow("photos", "share") && features.share,
      canServiceUpload: this.$config.feature("services") && this.$config.allow("services", "upload"),
      canManage: this.$config.allow("photos", "manage") && features.albums,
      canEdit: this.$config.allow("photos", "update") && features.edit,
      canEditAlbum: this.$config.allow("albums", "update") && features.albums,
      busy: false,
      config: this.$config.values,
      expanded: false,
      isAlbum: this.album && this.album.Type === "album",
      dialog: {
        archive: false,
        delete: false,
        album: false,
        share: false,
        webshare: false,
        move: false,
      },
      rtl: this.$isRtl,
    };
  },
  computed: {
    navigatorCanShare: function () {
      // Chrome has a limit on 10 items
      return canUseWebshareApi && this.selection.length <= 10;
    },
  },
  methods: {
    clearClipboard() {
      this.$clipboard.clear();
      this.expanded = false;
    },
    batchApprove() {
      if (this.busy || !this.canManage) {
        return;
      }

      this.busy = true;

      $api
        .post("batch/photos/approve", { photos: this.selection })
        .then(() => this.onApproved())
        .finally(() => {
          this.busy = false;
        });
    },
    onApproved() {
      $notify.success(this.$gettext("Selection approved"));
      this.clearClipboard();
    },
    archivePhotos() {
      if (!this.canArchive) {
        return;
      }

      if (!this.canDelete) {
        this.dialog.archive = true;
      } else {
        this.batchArchive();
      }
    },
    batchArchive() {
      if (this.busy || !this.canArchive) {
        return;
      }

      this.busy = true;
      this.dialog.archive = false;

      $api
        .post("batch/photos/archive", { photos: this.selection })
        .then(() => this.onArchived())
        .finally(() => {
          this.busy = false;
        });
    },
    onArchived() {
      $notify.success(this.$gettext("Selection archived"));
      this.clearClipboard();
    },
    deletePhotos() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = true;
    },
    batchDelete() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = false;

      $api.post("batch/photos/delete", { photos: this.selection }).then(() => this.onDeleted());
    },
    onDeleted() {
      $notify.success(this.$gettext("Permanently deleted"));
      this.clearClipboard();
    },
    batchPrivate() {
      $api.post("batch/photos/private", { photos: this.selection }).then(() => this.onPrivateSaved());
    },
    onPrivateSaved() {
      this.clearClipboard();
    },
    batchRestore() {
      $api.post("batch/photos/restore", { photos: this.selection }).then(() => this.onRestored());
    },
    onRestored() {
      $notify.success(this.$gettext("Selection restored"));
      this.clearClipboard();
    },
    addToAlbum(ppid) {
      if (!ppid || !this.canManage) {
        return;
      }

      if (this.busy) {
        return;
      }

      this.busy = true;
      this.dialog.album = false;

      $api
        .post(`albums/${ppid}/photos`, { photos: this.selection })
        .then(() => this.onAdded())
        .finally(() => {
          this.busy = false;
        });
    },
    onAdded() {
      this.clearClipboard();
    },
    removeFromAlbum() {
      if (!this.album) {
        this.$notify.error(this.$gettext("remove failed: unknown album"));
        return;
      }

      if (this.busy || !this.canManage) {
        return;
      }

      this.busy = true;

      const uid = this.album.UID;

      this.dialog.album = false;

      $api
        .delete(`albums/${uid}/photos`, { data: { photos: this.selection } })
        .then(() => this.onRemoved())
        .finally(() => {
          this.busy = false;
        });
    },
    onRemoved() {
      this.clearClipboard();
    },
    download() {
      if (this.busy || !this.canDownload) {
        return;
      }

      this.busy = true;

      switch (this.selection.length) {
        case 0:
          this.busy = false;
          return;
        case 1:
          new Photo()
            .find(this.selection[0])
            .then((p) => p.downloadAll())
            .finally(() => {
              this.busy = false;
            });
          break;
        default:
          $api
            .post("zip", { photos: this.selection })
            .then((r) => {
              this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken}`);
            })
            .finally(() => {
              this.busy = false;
            });
      }

      $notify.success(this.$gettext("Downloading…"));

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photos.zip");
    },
    onWebshareFailed() {
      this.busy = false;
      this.dialog.webshare = false;
      this.$notify.error(this.$gettext("sharing photos failed - showing download icon"));
    },
    webShare() {
      if (this.busy || this.selection.length == 0) {
        return;
      }
      this.busy = true;
      this.dialog.webshare = true;
      this.expanded = false;

      // try {
      //   let name = null;
      //
      //   new Photo().find(this.selection[0])
      //     .then((p) => {
      //       name = p.Name + '.' + p.Files[0].FileType;
      //       console.log(name);
      //
      //       fetch(`${$config.apiUri}/${p.getWebshareDownloadUrl()}`)
      //         .then(function (response) {
      //           return response.blob();
      //         })
      //         .then(function (blob) {
      //           let file = new File([blob], name, { type: blob.type });
      //           const shareData = {
      //             files: [file],
      //             title: "Share files",
      //           };
      //
      //           navigator
      //             .share(shareData)
      //             .then(() => console.log("shared"))
      //             .catch((e) => {
      //               this.$notify.error(e.message);
      //             });
      //         })
      //         .catch((e) => {
      //           this.$notify.error(e.message);
      //         })
      //         .finally(() => {});
      //     })
      //     .catch((e) => {
      //       this.$notify.error(e.message);
      //     });
      // } catch (e) {
      //   this.$notify.error(e.message);
      // }
    },
    edit() {
      // Open Edit Dialog
      this.$event.PubSub.publish("dialog.edit", { selection: this.selection, album: this.album, index: 0 });
    },
    onShared() {
      this.dialog.share = false;
      this.clearClipboard();
    },
  },
};
</script>
