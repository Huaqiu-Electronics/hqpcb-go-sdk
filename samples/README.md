# Go IPC samples（Arrow Flight · 逻辑管道名）

与 [`sdk/python/samples`](../../../python/samples) 对齐。客户端**只使用逻辑端点名**（默认 `hqpcb-svc`）。

## 启动服务端

```bash
cmake --build cmake-build-release --target pcb_IPC_skel_serve
./cmake-build-release/Runtime/IPC/pcb_IPC_skel_serve --pipe hqpcb-svc --role service
./cmake-build-release/Runtime/IPC/pcb_IPC_skel_serve --pipe hqpcb-gui --role gui
```

## 端点：headless / gui

与 Python/JS 同语义：`headless`/`gui`/`--headless`/`--gui`/`hqpcb-svc`/`hqpcb-gui`/`--endpoint=`/`PCB_IPC_ENDPOINT`。默认 headless。

- **board / ws 样例（04–10）**：`headless`（`hqpcb-svc`）与 **`gui`（`hqpcb-gui` / PCB_V5）** 均可；Gui 另有 interactive。
- 需重启已运行的 PCB_V5 后 `hqpcb-gui` 才会声明 `workspace=1`。

## 运行

```bash
cd sdk/go/pcb/samples
./run 01_hello_ping
./run 01_hello_ping gui
./run 10_objects_list_doget
./run 06_board_nets_vias headless /path/to/board.hqpcb
PCB_IPC_ENDPOINT=gui ./run 01_hello_ping
```

| 目录 | 内容 |
| --- | --- |
| `01_hello_ping` | ping / caps |
| `02_methods_list` | methods.list |
| `03_config` | config.get/set |
| `04_workspace_design` | ws.open + 设计租约写 |
| `05_complex_hierarchy_wire_length` | `OpenBoard` / `Tracks` / `LengthByNet` |
| `06_board_nets_vias` | `Nets` / `Vias` |
| `07_board_footprints` | `Footprints` / `Pads` / `Layers` |
| `08_board_zones_bbox` | `Zones` / `Drawings` / `BoundingBox` |
| `09_board_design_settings` | `DesignSettings` / `NetClasses` / … |
| `10_objects_list_doget` | DoGet `objects.list` / `query.table` |

默认板：`sdk/python/samples/complex_hierarchy.hqpcb`。
