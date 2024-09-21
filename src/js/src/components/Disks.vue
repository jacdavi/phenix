<template>
  <div class="content">
    <b-modal :active.sync="detailsModal.active" :on-cancel="() => detailsModal.active = false" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ detailsModal.disk.name }}</p>
        </header>
        <section class="modal-card-body">
          <p class="title is-5">Details</p>
          <dl>
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
                {{ i }}
              </div>
            </div>
          </div>

          <div v-if="roleAllowed('disks', 'post', detailsModal.disk.name)" class="actions">
            <hr>
            <p class="title is-5">Actions</p>
            <b-button type="is-text" expanded @click="snapshotDisk(detailsModal.disk.fullPath)">
              <b>Snapshot</b> - Creates a new image backed by this image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded @click="commitDisk(detailsModal.disk.fullPath)">
              <b>Commit</b> - Commits change in this image to its backing image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded>
              <b>Rebase</b> - Updates image and rebases onto a different backing image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded>
              <b>Set Backing</b> - Sets the backing file without changing image
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded>
              <b>Clone</b>
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded>
              <b>Download</b>
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded>
              <b>Rename</b>
            </b-button>
            <hr class="action-separator">
            <b-button type="is-text" expanded>
              <b>Delete</b>
            </b-button>
          </div>
        </section>
      </div>

<!-- <b-dropdown-item aria-role="listitem">Snapshot</b-dropdown-item>
<b-dropdown-item aria-role="listitem">Commit</b-dropdown-item>
<b-dropdown-item aria-role="listitem">Rebase</b-dropdown-item>
<b-dropdown-item aria-role="listitem">Clone</b-dropdown-item>
<hr class="dropdown-divider">
<b-dropdown-item aria-role="listitem">Download</b-dropdown-item>
<b-dropdown-item aria-role="listitem">Rename</b-dropdown-item>
<b-dropdown-item aria-role="listitem">Delete</b-dropdown-item>
 -->
    </b-modal>
    <template>
      <hr>
      <b-field grouped position="is-right">
        <b-tooltip label="Refresh List" type="is-light is-left">
          <button class="button is-light" @click="updateDisks">
            <b-icon icon="refresh"></b-icon>
          </button>
        </b-tooltip>
        <b-tooltip v-if="roleAllowed('disks', 'post')" label="Upload a disk" type="is-light is-left">
          <button class="button is-light" style="margin-left: 8px;">
            <b-icon icon="upload"></b-icon>
          </button>
        </b-tooltip>
      </b-field>
      <div>
        <b-table hoverable @click="rowClick" :row-class="(r, i) => 'is-clickable'" :data="disks"
          :paginated="table.isPaginated" :per-page="table.perPage" :current-page.sync="table.currentPage"
          :pagination-simple="table.isPaginationSimple" :pagination-size="table.paginationSize"
          :default-sort-direction="table.defaultSortDirection" :loading="isWaiting" default-sort="name">
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
  },

  methods: {
    updateDisks() {
      this.detailsModal.active = false
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
    actionWrapper(httpPath) {
      this.$http.post(httpPath).then(
            _ => this.updateDisks(),
            err => this.errorNotification(err)
          )
    },
    commitDisk(path) {
      console.log(path)
      this.$buefy.dialog.confirm({
        message: "Are you sure you want to commit this disk? The disk will remain unchanged, but its backing image will contain all data from this disk.",
        onConfirm: this.actionWrapper(`disks/commit?path=${path}`)
      })
    },
    snapshotDisk(path) {
      console.log(path)
      this.$buefy.dialog.prompt({
        message: "Are you sure you want to snapshot this disk?",
        inputAttrs: {
          type: "text",
          "placeholder": "New image name"
        },
        onConfirm: (value) => this.actionWrapper(`disks/snapshot?src=${path}&dst=${value}`)
      })
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
      disks: [],
      isWaiting: false,
      detailsModal: {
        active: false,
        disk: {}
      },
    }
  }
}
</script>

<style scoped>
.b-tooltip:after {
  white-space: pre !important;
}

.modal-card-body {
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

.actions > button {
  text-align: start;
  color: dimgray;
  text-decoration: none;
  display: inline;
}
</style>