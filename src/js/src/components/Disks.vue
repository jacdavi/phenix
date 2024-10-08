<template>
  <div class="content">
    <!-- DETAILS MODAL -->
    <b-modal :active.sync="detailsModal.active" :on-cancel="() => detailsModal.active = false" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ detailsModal.disk.name }}</p>
        </header>
        <section class="modal-card-body">
          <p class="title is-5">Details</p>
          <dl>
            <div>
              <dt>Host:</dt>
              <dd>{{ detailsModal.disk.host }}</dd>
            </div>
            <div>
              <dt>Full Path:</dt>
              <dd>{{ detailsModal.disk.fullPath }}</dd>
            </div>
            <div>
              <dt>Kind:</dt>
              <dd>{{ detailsModal.disk.kind }}</dd>
            </div>
            <div>
              <dt>Size on Disk:</dt>
              <dd>{{ detailsModal.disk.size }}</dd>
            </div>
            <div>
              <dt>Virtual Size:</dt>
              <dd>{{ detailsModal.disk.virtualSize }}</dd>
            </div>
            <div>
              <dt>Experiment:</dt>
              <dd>{{ detailsModal.disk.experiment || "N/A" }}</dd>
            </div>
            <div>
              <dt>In Use:</dt>
              <dd>{{ detailsModal.disk.inUse }}</dd>
            </div>
          </dl>
          <div v-if="detailsModal.disk.backingImages && detailsModal.disk.backingImages.length > 0">
            <hr>
            <p class="title is-5">Backing Chain</p>
            <div style="text-align: center;">
              <b>{{ detailsModal.disk.name }}</b>
              <div v-for="i in detailsModal.disk.backingImages">
                &darr;<br>
                <a @click="detailsModal.disk = disks.find(d => d.name==i)" >{{  i  }}</a>
              </div>
            </div>
          </div>

          <div class="actions">
            <hr>
            <p class="title is-5">Actions</p>
            <b-button type="is-text" expanded @click="snapshotDisk(detailsModal.disk.fullPath)"
              :disabled="shouldDisableAction('snapshot')">
              <b>Snapshot</b> - Creates a new image backed by this image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="() => commitModal.active = true"
              :disabled="shouldDisableAction('commit')">
              <b>Commit</b> - Commits change in this image to its backing image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="() => rebaseModal.active = true"
              :disabled="shouldDisableAction('rebase')">
              <b>Rebase</b> - Updates image and rebases onto a different backing image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="cloneDisk(detailsModal.disk.fullPath)"
              :disabled="shouldDisableAction('clone')">
              <b>Clone</b> - Creates a copy of the disk file
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="renameDisk(detailsModal.disk.fullPath)"
              :disabled="shouldDisableAction('rename')">
              <b>Rename</b>
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="deleteDisk(detailsModal.disk.fullPath)"
              :disabled="shouldDisableAction('delete')">
              <b>Delete</b>
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="downloadDisk(detailsModal.disk.fullPath)"
              :disabled="shouldDisableAction('download')">
              <b>Download</b>
            </b-button>
          </div>
        </section>
      </div>
    </b-modal>
    <!-- REBASE MODAL -->
    <b-modal :active.sync="rebaseModal.active" :on-cancel="() => rebaseModal.active = false" has-modal-card>
      <div class="modal-card" style="max-width: 460px;">
        <section class="modal-card-body">
          Are you sure you want to rebase this image onto a different backing image?<br>
          By default changes between the old and new backing images will be written to this image.
          Selecting "None" for the backing image will cause the image to become independent.<br>
          Selecting "Change Reference Only" will only change the backing image name without updating files.
          <b-select placeholder="New Backing Image" v-model="rebaseModal.dst" style="margin-bottom: 8px; margin-top: 16px;">
            <option value="">None</option>
            <option v-for="d in disks" :value="d.fullPath">{{ d.name }}</option>
          </b-select>
          <b-checkbox v-model="rebaseModal.unsafe">Change reference only</b-checkbox>
        </section>
        <footer class="modal-card-foot" style="justify-content: flex-end;">
          <b-button label="Close" @click="() => rebaseModal.active = false" />
          <b-button label="OK" type="is-primary"
            @click="() => rebaseDisk(detailsModal.disk.fullPath, rebaseModal.dst, rebaseModal.unsafe)" />
        </footer>
      </div>
    </b-modal>
    <!-- COMMIT MODAL -->
    <b-modal :active.sync="commitModal.active" :on-cancel="() => commitModal.active = false" has-modal-card>
      <div class="modal-card" style="max-width: 460px;">
        <section class="modal-card-body">
          Are you sure you want to commit the changes in this disk to its parent?<br>
          By default this disk is left unchanged, but you may select to delete if it's no longer needed.
          <b-field style="margin-top: 16px;">
            <b-checkbox v-model="commitModal.delete">Delete this disk after commit</b-checkbox>
          </b-field>
        </section>
        <footer class="modal-card-foot" style="justify-content: flex-end;">
          <b-button label="Close" @click="() => commitModal.active = false" />
          <b-button label="OK" type="is-primary"
            @click="() => commitDisk(detailsModal.disk.fullPath, commitModal.delete)" />
        </footer>
      </div>
    </b-modal>
    <!-- CONTENT -->
    <template>
      <hr>
      <b-field grouped position="is-right">
        <b-field>
          <b-autocomplete v-model="filterString" placeholder="Find a disk" icon="search"
          @select="option => selected = option" :data="filteredDisks.map(d => d.name)" style="width: 512px;">
          </b-autocomplete>
          <p class='control'>
            <button class='button' style="color:#686868" @click="filterString = ''">
              <b-icon icon="window-close"></b-icon>
            </button>
          </p>
        </b-field>

        <b-tooltip label="Refresh List" type="is-light is-left">
          <button class="button is-light" @click="updateDisks">
            <b-icon icon="refresh"></b-icon>
          </button>
        </b-tooltip>
        <b-tooltip v-if="roleAllowed('disks', 'upload')" label="Upload a disk" type="is-light is-left">
          <b-upload class="file-label" style="margin-left: 8px;" @input="uploadDisk"
            :disabled="currentUploadProgress != null">
            <span class="file-cta">
              <b-icon v-if="currentUploadProgress == null" icon="upload"></b-icon>
              <p v-else style="width: 32px;"> {{ currentUploadProgress }}% </p>
            </span>
          </b-upload>
        </b-tooltip>
      </b-field>
      <div>
        <b-table hoverable @click="rowClick" :row-class="(r, i) => 'is-clickable'" :data="filteredDisks"
          :paginated="table.isPaginated" :per-page="table.perPage" :current-page.sync="table.currentPage"
          :pagination-simple="table.isPaginationSimple" :pagination-size="table.paginationSize"
          :default-sort-direction="table.defaultSortDirection" :loading="isWaiting" default-sort="name">
          <template slot="empty">
            <section class="section">
              <div class="content has-text-white has-text-centered">
                No Disks Found
              </div>
            </section>
          </template>
          <b-table-column field="name" label="Name" v-slot="props">
            {{ props.row.name }}
          </b-table-column>
          <b-table-column field="kind" label="Kind" v-slot="props">
            {{ props.row.kind }}
          </b-table-column>
          <b-table-column field="inUse" label="In Use" centered v-slot="props">
            <b-icon v-if="props.row.inUse" icon="play-circle" size="is-small" />
          </b-table-column>
          <b-table-column field="size" label="Size" v-slot="props">
            {{ props.row.size | fileSize }}
          </b-table-column>
        </b-table>
        <br>
        <b-field v-if="paginationNeeded" grouped position="is-right">
          <div class="control is-flex">
            <b-switch v-model="table.isPaginated" size="is-small" type="is-light"
              @input="changePaginate()">Paginate</b-switch>
          </div>
        </b-field>
      </div>
    </template>
  </div>
</template>

<script>
export default {

  beforeDestroy() {
  },

  async created() {
    this.updateDisks();
    this.table.isPaginated = localStorage.getItem(localStorage.getItem('user') + '.lastPaginate') == 'true';

  },

  computed: {
    paginationNeeded() {
      return this.disks.length > this.table.perPage
    },
    filteredDisks() {
      return this.disks == null ? [] : this.disks.filter((disk) => disk.name.toLowerCase().indexOf(this.filterString.toLowerCase()) >= 0)
    }
  },

  methods: {
    updateDisks() {
      this.detailsModal.active = false
      this.rebaseModal = {
        active: false,
        unsafe: false,
        dst: ""
      }
      this.commitModal = {
        active: false,
        delete: false
      }
      this.isWaiting = true
      this.$http.get('disks').then(
        response => {
          response.json().then(
            state => {
              console.log(state)
              this.disks = []

              for (let i = 0; i < state.disks.length; i++) {
                this.disks.push(state.disks[i]);
              }

              this.disks.sort()
              this.isWaiting = false;
            }

          );
        }, err => {
          this.errorNotification(err);
        }
      );
    },
    rowClick(row) {
      console.log(row)
      this.detailsModal.disk = row
      this.detailsModal.active = true
    },
    changePaginate() {
      var user = localStorage.getItem('user');
      localStorage.setItem(user + '.lastPaginate', this.table.isPaginated);
    },

    shouldDisableAction(action) {
      let disk = this.detailsModal.disk
      switch (action) {
        case "snapshot":
          return disk.inUse || disk.kind != "VM" || !this.roleAllowed('disks', 'create')
        case "commit":
          return disk.inUse || (disk.backingImages && disk.backingImages.length == 0) || disk.kind != "VM"
        case "rebase":
          return disk.inUse || disk.kind != "VM" || !this.roleAllowed('disks', 'update', disk.name)
        case "delete":
          return disk.inUse || !this.roleAllowed('disks', 'delete', disk.name)
        case "rename":
          return disk.inUse || !this.roleAllowed('disks', 'update', disk.name)
        case "clone":
          return !this.roleAllowed('disks', 'create')
        case "download":
          return !this.roleAllowed('disks', 'get', disk.name)
        default:
          return false
      }
    },
    actionWrapper(httpPath, method='post') {
      this.$http[method](httpPath).then(
        _ => this.updateDisks(),
        err => this.errorNotification(err)
      )
    },
    commitDisk(path, deleteOnSuccess) {
      console.log(path)
      this.$http.post(`disks/commit?disk=${path}`).then(
        _ => {
          if (deleteOnSuccess) {
            this.actionWrapper(`disks?disk=${path}`, 'delete')
          }
          else {
            this.updateDisks()
          }
        },
        err => errorNotification(err)
      )
    },
    snapshotDisk(path) {
      console.log(path)
      this.$buefy.dialog.prompt({
        message: "Are you sure you want to snapshot this disk? This will create a new disk backed by this image.",
        inputAttrs: {
          type: "text",
          placeholder: "New image name"
        },
        onConfirm: (value) => this.actionWrapper(`disks/snapshot?disk=${path}&new=${value}`)
      })
    },
    rebaseDisk(path, dst, unsafe) {
      this.actionWrapper(`disks/rebase?disk=${path}&backing=${dst}&unsafe=${unsafe}`)
    },
    cloneDisk(path) {
      console.log(path)
      this.$buefy.dialog.confirm({
        message: "Are you sure you want to clone this disk?.",
        inputAttrs: {
          type: "text",
          placeholder: "New image name"
        },
        onConfirm: (value) => this.actionWrapper(`disks/clone?disk=${path}&new=${value}`)
      })
    },
    renameDisk(path) {
      console.log(path)
      this.$buefy.dialog.confirm({
        message: 'Are you sure you want to rename this disk? <b class="has-text-danger">If this disk backs others, they must be rebased to use the new name</b>',
        inputAttrs: {
          type: "text",
          placeholder: "New name"
        },
        onConfirm: (value) => this.actionWrapper(`disks/rename?disk=${path}&new=${value}`)
      })
    },
    deleteDisk(path) {
      console.log(path)
      this.$buefy.dialog.confirm({
        message: 'Are you sure you want to delete this disk? <b class="has-text-danger">If this disk backs others, they will become invalid</b>',
        onConfirm: () => this.actionWrapper(`disks?disk=${path}`, 'delete')
      })
    },
    downloadDisk(path) {
      this.$buefy.dialog.confirm({
        message: 'Are you sure you want to download this disk?',
        onConfirm: () => window.open(`${process.env.BASE_URL}api/v1/disks/download?token=${this.$store.state.token}&disk=${encodeURIComponent(path)}`, '_blank')
      })
    },
    uploadDisk(file) {
      let formData = new FormData();
      formData.append('file', file);
      this.currentUploadProgress = 0;
      console.log(file.name)
      this.$http.post(`disks`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        uploadProgress: (event) => {
          this.currentUploadProgress = Math.round(event.loaded / event.total * 100);
        }
      }).then(_ => {
        this.currentUploadProgress = null;
        this.updateDisks()
      }, err => {
        this.errorNotification(`Error uploading: ${err.body}`)
        this.currentUploadProgress = null;
      });
    }
  },

  data() {
    return {
      table: {
        striped: true,
        isPaginated: false,
        isPaginationSimple: true,
        paginationSize: 'is-small',
        defaultSortDirection: 'asc',
        currentPage: 1,
        perPage: 10
      },
      currentUploadProgress: null,
      disks: [],
      filterString: "",
      isWaiting: false,
      detailsModal: {
        active: false,
        disk: {}
      },
      rebaseModal: {
        active: false,
        unsafe: false,
        dst: ""
      },
      commitModal: {
        active: false,
        delete: false,
      }
    }
  }
}
</script>

<style scoped>
.b-tooltip:after {
  white-space: pre !important;
}

.modal-card-body, .modal-card-body a {
  color: black !important;
}

dl {
  display: table;
}

dl>div {
  display: table-row;
}

dl>div>dt,
dl>div>dd {
  display: table-cell;
  padding: 0.25em;
}

dl>div>dt {
  font-weight: bold;
  width: 20%;
}

hr {
  margin: 4px 0px;
}

.action-button {
  color: dimgray;
  padding: 8px;
  cursor: pointer !important;
}

.action-button:hover {
  background-color: #DDD;
}

.action-separator {
  margin: 0 8px;
}

.actions>button {
  text-align: start;
  color: dimgray;
  text-decoration: none;
  display: inline;
}

.file-cta,
.file-cta>p,
.file-cta:hover {
  border: none;
  background-color: #686868;
  color: whitesmoke !important;
}

div.autocomplete>>>a.dropdown-item {
  color: #383838 !important;
}
</style>