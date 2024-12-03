export interface ServerStatus {
  cpu: number[];
  host: Host;
  mem: Mem;
  disk: Disk;
  nic: Nic[];
}

export interface Host {
  hostname: string;
  uptime: number;
  bootTime: number;
  procs: number;
  os: string;
  platform: string;
  platformFamily: string;
  platformVersion: string;
  kernelVersion: string;
  kernelArch: string;
  virtualizationSystem: string;
  virtualizationRole: string;
  hostId: string;
}

export interface Mem {
  total: number;
  available: number;
  used: number;
  usedPercent: number;
  free: number;
  active: number;
  inactive: number;
  wired: number;
  laundry: number;
  buffers: number;
  cached: number;
  writeBack: number;
  dirty: number;
  writeBackTmp: number;
  shared: number;
  slab: number;
  sreclaimable: number;
  sunreclaim: number;
  pageTables: number;
  swapCached: number;
  commitLimit: number;
  committedAS: number;
  highTotal: number;
  highFree: number;
  lowTotal: number;
  lowFree: number;
  swapTotal: number;
  swapFree: number;
  mapped: number;
  vmallocTotal: number;
  vmallocUsed: number;
  vmallocChunk: number;
  hugePagesTotal: number;
  hugePagesFree: number;
  hugePagesRsvd: number;
  hugePagesSurp: number;
  hugePageSize: number;
  anonHugePages: number;
}

export interface Disk {
  path: string;
  fstype: string;
  total: number;
  free: number;
  used: number;
  usedPercent: number;
  inodesTotal: number;
  inodesUsed: number;
  inodesFree: number;
  inodesUsedPercent: number;
}

export interface Nic {
  name: string;
  bytesSent: number;
  bytesRecv: number;
  packetsSent: number;
  packetsRecv: number;
  errin: number;
  errout: number;
  dropin: number;
  dropout: number;
  fifoin: number;
  fifoout: number;
}
