```mermaid
flowchart TB
  %% ========= UI 层 =========
  subgraph UI_API [Web UI / API]
    direction TB
    WebUI[/"🌐 Web UI/REST API"/]
  end

  %% ========= 控制层 =========
  subgraph Controller[Helios Controller (x86)]
    direction TB

    Scheduler[🧠 调度器 / 策略引擎<br/>(Scheduler / Policy Engine)]
    MetadataDB[📦 模型/数据元数据中心<br/>(Metadata Database)]
    SharedL3[📶 L3 共享缓存 (Shared Cache)]

    Scheduler --> MetadataDB
    MetadataDB --> SharedL3
  end

  %% ========= RPC 链接 =========
  WebUI -->|HTTP / REST| Controller
  Controller -->|Cap'n Proto RPC<br/>over 1000Gb/s RoCE| Agents

  %% ========= 多节点 Agent 层 =========
  subgraph Agents [AC922 GPU Nodes]
    direction LR

    %% --- Server 1 ---
    subgraph S1[AC922 Server 1]
      Agent1[🛰️ Helios Agent<br/>+ CLI]
      subgraph NUMA0_S1[NUMA 0 Container]
        LM0[🤖 Megatron-LM #0]
        LM0_Mem[L2: Host Memory<br/>L1: GPU VRAM]
      end
      subgraph NUMA1_S1[NUMA 1 Container]
        LM1[🤖 Megatron-LM #1]
        LM1_Mem[L2: Host Memory<br/>L1: GPU VRAM]
      end
      Agent1 --> NUMA0_S1
      Agent1 --> NUMA1_S1
      LM0 --> LM0_Mem
      LM1 --> LM1_Mem
    end

    %% --- Server 2 ---
    subgraph S2[AC922 Server 2]
      Agent2[🛰️ Helios Agent<br/>+ CLI]
      subgraph NUMA0_S2[NUMA 0 Container]
        LM2[🤖 Megatron-LM #2]
        LM2_Mem[L2: Host Memory<br/>L1: GPU VRAM]
      end
      subgraph NUMA1_S2[NUMA 1 Container]
        LM3[🤖 Megatron-LM #3]
        LM3_Mem[L2: Host Memory<br/>L1: GPU VRAM]
      end
      Agent2 --> NUMA0_S2
      Agent2 --> NUMA1_S2
      LM2 --> LM2_Mem
      LM3 --> LM3_Mem
    end

    %% 可扩展节点
    Dots[⋯更多 Server]

  end
```
