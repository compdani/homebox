import { LabelsApi } from "./classes/labels";
import { LocationsApi } from "./classes/locations";
import { ProductsApi } from "./classes/products";
import { ItemsApi } from "./classes/items";
import { GroupApi } from "./classes/group";
import { UserApi } from "./classes/users";
import { ActionsAPI } from "./classes/actions";
import { StatsAPI } from "./classes/stats";
import { AssetsApi } from "./classes/assets";
import { ReportsAPI } from "./classes/reports";
import { NotifiersAPI } from "./classes/notifiers";
import type { Requests } from "~~/lib/requests";

export class UserClient {
  locations: LocationsApi;
  labels: LabelsApi;
  products: ProductsApi;
  items: ItemsApi;
  group: GroupApi;
  user: UserApi;
  actions: ActionsAPI;
  stats: StatsAPI;
  assets: AssetsApi;
  reports: ReportsAPI;
  notifiers: NotifiersAPI;

  constructor(requests: Requests, attachmentToken: string) {
    this.locations = new LocationsApi(requests);
    this.labels = new LabelsApi(requests);
    this.products = new ProductsApi(requests);
    this.items = new ItemsApi(requests, attachmentToken);
    this.group = new GroupApi(requests);
    this.user = new UserApi(requests);
    this.actions = new ActionsAPI(requests);
    this.stats = new StatsAPI(requests);
    this.assets = new AssetsApi(requests);
    this.reports = new ReportsAPI(requests);
    this.notifiers = new NotifiersAPI(requests);

    Object.freeze(this);
  }
}
