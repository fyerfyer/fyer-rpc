"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[154],{4393:(e,r,n)=>{n.r(r),n.d(r,{assets:()=>d,contentTitle:()=>l,default:()=>h,frontMatter:()=>c,metadata:()=>s,toc:()=>o});const s=JSON.parse('{"id":"protocol/protocol-intro","title":"Protocol","description":"fyerrpc\u6846\u67b6\u91c7\u7528\u7cbe\u5fc3\u8bbe\u8ba1\u7684\u4e8c\u8fdb\u5236\u534f\u8bae\u6765\u4fdd\u8bc1\u9ad8\u6548\u3001\u53ef\u9760\u7684RPC\u901a\u4fe1\u3002\u672c\u6587\u6863\u8be6\u7ec6\u4ecb\u7ecd\u4e86fyerrpc\u7684\u534f\u8bae\u683c\u5f0f\u3001\u7ec4\u6210\u90e8\u5206\u4ee5\u53ca\u5982\u4f55\u6269\u5c55\u548c\u81ea\u5b9a\u4e49\u534f\u8bae\u3002","source":"@site/docs/protocol/protocol-intro.md","sourceDirName":"protocol","slug":"/protocol/protocol-intro","permalink":"/fyer-rpc/docs/protocol/protocol-intro","draft":false,"unlisted":false,"editUrl":"https://github.com/fyerfyer/fyer-rpc/tree/main/docs/protocol/protocol-intro.md","tags":[],"version":"current","frontMatter":{},"sidebar":"tutorialSidebar","previous":{"title":"fyerrpc","permalink":"/fyer-rpc/docs/intro"},"next":{"title":"Client","permalink":"/fyer-rpc/docs/user-guide/client/"}}');var i=n(4848),a=n(8453);const c={},l="Protocol",d={},o=[{value:"\u6d88\u606f\u683c\u5f0f",id:"\u6d88\u606f\u683c\u5f0f",level:2},{value:"\u6574\u4f53\u7ed3\u6784",id:"\u6574\u4f53\u7ed3\u6784",level:3},{value:"\u5143\u6570\u636e (Metadata)",id:"\u5143\u6570\u636e-metadata",level:3},{value:"\u534f\u8bae\u5934 (Header)",id:"\u534f\u8bae\u5934-header",level:2},{value:"\u5934\u90e8\u7ed3\u6784",id:"\u5934\u90e8\u7ed3\u6784",level:3},{value:"\u5934\u90e8\u5b57\u6bb5\u8be6\u89e3",id:"\u5934\u90e8\u5b57\u6bb5\u8be6\u89e3",level:3},{value:"\u6d88\u606f\u7f16\u89e3\u7801",id:"\u6d88\u606f\u7f16\u89e3\u7801",level:2},{value:"\u9ed8\u8ba4\u534f\u8bae\u5b9e\u73b0",id:"\u9ed8\u8ba4\u534f\u8bae\u5b9e\u73b0",level:3},{value:"\u5e8f\u5217\u5316",id:"\u5e8f\u5217\u5316",level:2},{value:"\u5185\u7f6e\u5e8f\u5217\u5316\u5b9e\u73b0",id:"\u5185\u7f6e\u5e8f\u5217\u5316\u5b9e\u73b0",level:3},{value:"JSON \u5e8f\u5217\u5316",id:"json-\u5e8f\u5217\u5316",level:4},{value:"Protobuf \u5e8f\u5217\u5316",id:"protobuf-\u5e8f\u5217\u5316",level:4},{value:"\u81ea\u5b9a\u4e49\u534f\u8bae",id:"\u81ea\u5b9a\u4e49\u534f\u8bae",level:2},{value:"\u6269\u5c55\u9ed8\u8ba4\u534f\u8bae",id:"\u6269\u5c55\u9ed8\u8ba4\u534f\u8bae",level:3},{value:"\u5b9e\u73b0\u5168\u65b0\u534f\u8bae",id:"\u5b9e\u73b0\u5168\u65b0\u534f\u8bae",level:3},{value:"\u6ce8\u518c\u81ea\u5b9a\u4e49\u5e8f\u5217\u5316\u5668",id:"\u6ce8\u518c\u81ea\u5b9a\u4e49\u5e8f\u5217\u5316\u5668",level:3},{value:"\u534f\u8bae\u4f7f\u7528\u793a\u4f8b",id:"\u534f\u8bae\u4f7f\u7528\u793a\u4f8b",level:2},{value:"\u57fa\u672c\u4f7f\u7528",id:"\u57fa\u672c\u4f7f\u7528",level:3},{value:"\u4f7f\u7528\u81ea\u5b9a\u4e49\u534f\u8bae",id:"\u4f7f\u7528\u81ea\u5b9a\u4e49\u534f\u8bae",level:3},{value:"\u534f\u8bae\u8bbe\u8ba1\u8003\u91cf",id:"\u534f\u8bae\u8bbe\u8ba1\u8003\u91cf",level:2},{value:"\u6027\u80fd\u4f18\u5316",id:"\u6027\u80fd\u4f18\u5316",level:3},{value:"\u53ef\u6269\u5c55\u6027",id:"\u53ef\u6269\u5c55\u6027",level:3},{value:"\u53ef\u8c03\u8bd5\u6027",id:"\u53ef\u8c03\u8bd5\u6027",level:3},{value:"\u6700\u4f73\u5b9e\u8df5",id:"\u6700\u4f73\u5b9e\u8df5",level:2},{value:"\u534f\u8bae\u9009\u62e9",id:"\u534f\u8bae\u9009\u62e9",level:3},{value:"\u534f\u8bae\u6269\u5c55",id:"\u534f\u8bae\u6269\u5c55",level:3},{value:"\u5e8f\u5217\u5316\u9009\u62e9",id:"\u5e8f\u5217\u5316\u9009\u62e9",level:3}];function t(e){const r={code:"code",h1:"h1",h2:"h2",h3:"h3",h4:"h4",header:"header",li:"li",ol:"ol",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(r.header,{children:(0,i.jsx)(r.h1,{id:"protocol",children:"Protocol"})}),"\n",(0,i.jsx)(r.p,{children:"fyerrpc\u6846\u67b6\u91c7\u7528\u7cbe\u5fc3\u8bbe\u8ba1\u7684\u4e8c\u8fdb\u5236\u534f\u8bae\u6765\u4fdd\u8bc1\u9ad8\u6548\u3001\u53ef\u9760\u7684RPC\u901a\u4fe1\u3002\u672c\u6587\u6863\u8be6\u7ec6\u4ecb\u7ecd\u4e86fyerrpc\u7684\u534f\u8bae\u683c\u5f0f\u3001\u7ec4\u6210\u90e8\u5206\u4ee5\u53ca\u5982\u4f55\u6269\u5c55\u548c\u81ea\u5b9a\u4e49\u534f\u8bae\u3002"}),"\n",(0,i.jsx)(r.h2,{id:"\u6d88\u606f\u683c\u5f0f",children:"\u6d88\u606f\u683c\u5f0f"}),"\n",(0,i.jsx)(r.p,{children:"fyerrpc\u91c7\u7528\u7b80\u5355\u9ad8\u6548\u7684\u4e8c\u8fdb\u5236\u683c\u5f0f\uff0c\u4e00\u4e2a\u5b8c\u6574\u7684RPC\u6d88\u606f\u7531\u4e09\u90e8\u5206\u7ec4\u6210\uff1a\u534f\u8bae\u5934(Header)\u3001\u5143\u6570\u636e(Metadata)\u548c\u6d88\u606f\u4f53(Payload)\u3002"}),"\n",(0,i.jsx)(r.h3,{id:"\u6574\u4f53\u7ed3\u6784",children:"\u6574\u4f53\u7ed3\u6784"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{children:"+------------------+\r\n|     Header      |  \u6d88\u606f\u5934\u90e8\uff08\u56fa\u5b9a22\u5b57\u8282\uff09\r\n+------------------+\r\n|    Metadata     |  \u5143\u6570\u636e(\u53ef\u53d8\u957f\u5ea6)\uff0c\u5305\u542b\u670d\u52a1\u540d\u3001\u65b9\u6cd5\u540d\u7b49\u4fe1\u606f\r\n+------------------+\r\n|    Payload      |  \u6d88\u606f\u4f53(\u53ef\u53d8\u957f\u5ea6)\uff0c\u5305\u542b\u8bf7\u6c42\u53c2\u6570\u6216\u54cd\u5e94\u7ed3\u679c\r\n+------------------+\n"})}),"\n",(0,i.jsx)(r.p,{children:"\u5728Go\u4ee3\u7801\u4e2d\uff0c\u6d88\u606f\u7ed3\u6784\u5b9a\u4e49\u5982\u4e0b\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"type Message struct {\r\n    Header   Header    // \u6d88\u606f\u5934\u90e8\r\n    Metadata *Metadata // \u5143\u6570\u636e\r\n    Payload  []byte    // \u6d88\u606f\u4f53\r\n}\n"})}),"\n",(0,i.jsx)(r.h3,{id:"\u5143\u6570\u636e-metadata",children:"\u5143\u6570\u636e (Metadata)"}),"\n",(0,i.jsx)(r.p,{children:"\u5143\u6570\u636e\u5305\u542b\u4e86RPC\u8c03\u7528\u7684\u6838\u5fc3\u4fe1\u606f\uff0c\u5982\u670d\u52a1\u540d\u79f0\u3001\u65b9\u6cd5\u540d\u79f0\u3001\u9519\u8bef\u4fe1\u606f\u7b49\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"type Metadata struct {\r\n    ServiceName string            // \u670d\u52a1\u540d\u79f0\r\n    MethodName  string            // \u65b9\u6cd5\u540d\u79f0\r\n    Error       string            // \u9519\u8bef\u4fe1\u606f(\u4ec5\u54cd\u5e94\u6d88\u606f\u4f7f\u7528)\r\n    Extra       map[string]string // \u989d\u5916\u7684\u5143\u6570\u636e\uff0c\u5982trace_id\u7b49\r\n}\n"})}),"\n",(0,i.jsxs)(r.p,{children:["\u5143\u6570\u636e\u652f\u6301\u7528\u6237\u81ea\u5b9a\u4e49\u6269\u5c55\u5b57\u6bb5\uff0c\u53ef\u4ee5\u901a\u8fc7",(0,i.jsx)(r.code,{children:"Extra"}),"\u5b57\u6bb5\u6dfb\u52a0\u94fe\u8def\u8ffd\u8e2aID\u3001\u8ba4\u8bc1\u4fe1\u606f\u7b49\u9644\u52a0\u6570\u636e\u3002"]}),"\n",(0,i.jsx)(r.h2,{id:"\u534f\u8bae\u5934-header",children:"\u534f\u8bae\u5934 (Header)"}),"\n",(0,i.jsx)(r.p,{children:"\u534f\u8bae\u5934\u662ffyerrpc\u6d88\u606f\u7684\u56fa\u5b9a\u90e8\u5206\uff0c\u5305\u542b\u4e86\u5904\u7406\u6d88\u606f\u6240\u9700\u7684\u6240\u6709\u63a7\u5236\u4fe1\u606f\uff0c\u91c7\u7528\u56fa\u5b9a\u957f\u5ea6\u7684\u4e8c\u8fdb\u5236\u683c\u5f0f\u3002"}),"\n",(0,i.jsx)(r.h3,{id:"\u5934\u90e8\u7ed3\u6784",children:"\u5934\u90e8\u7ed3\u6784"}),"\n",(0,i.jsx)(r.p,{children:"\u534f\u8bae\u5934\u603b\u517122\u4e2a\u5b57\u8282\uff0c\u6309\u5b57\u8282\u5212\u5206\u7684\u683c\u5f0f\u5982\u4e0b\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{children:"+-----------------------------------------------+\r\n|  magic number   |  version    |  msg type     |\r\n+-----------------------------------------------+\r\n|  2 bytes        |  1 byte     |  1 byte       |\r\n+-----------------------------------------------+\r\n|  compress type  |  serial type|  message id   |\r\n+-----------------------------------------------+\r\n|  1 byte         |  1 byte     |  8 bytes      |\r\n+-----------------------------------------------+\r\n|  metadata size  |  payload size               |\r\n+-----------------------------------------------+\r\n|  4 bytes        |  4 bytes                    |\r\n+-----------------------------------------------+\n"})}),"\n",(0,i.jsx)(r.p,{children:"\u5728Go\u4e2d\u7684\u5b9a\u4e49\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"type Header struct {\r\n    MagicNumber       uint16 // \u9b54\u6570\uff0c\u7528\u4e8e\u6821\u9a8c\u62a5\u6587\r\n    Version           uint8  // \u534f\u8bae\u7248\u672c\u53f7\r\n    MessageType       uint8  // \u6d88\u606f\u7c7b\u578b(\u8bf7\u6c42/\u54cd\u5e94)\r\n    CompressType      uint8  // \u538b\u7f29\u7c7b\u578b\r\n    SerializationType uint8  // \u5e8f\u5217\u5316\u7c7b\u578b\r\n    MessageID         uint64 // \u6d88\u606fID\uff0c\u7528\u4e8e\u591a\u8def\u590d\u7528\r\n    MetadataSize      uint32 // \u5143\u6570\u636e\u957f\u5ea6\r\n    PayloadSize       uint32 // \u6d88\u606f\u4f53\u957f\u5ea6\r\n}\n"})}),"\n",(0,i.jsx)(r.h3,{id:"\u5934\u90e8\u5b57\u6bb5\u8be6\u89e3",children:"\u5934\u90e8\u5b57\u6bb5\u8be6\u89e3"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u9b54\u6570 (Magic Number)"})," - 2\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsxs)(r.li,{children:["\u56fa\u5b9a\u503c\uff1a",(0,i.jsx)(r.code,{children:"0x3f3f"})]}),"\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u5feb\u901f\u6821\u9a8c\u662f\u5426\u4e3a\u6709\u6548\u7684fyerrpc\u6d88\u606f\uff0c\u907f\u514d\u5904\u7406\u9519\u8bef\u7684\u6d88\u606f"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u7248\u672c (Version)"})," - 1\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsxs)(r.li,{children:["\u5f53\u524d\u503c\uff1a",(0,i.jsx)(r.code,{children:"0x01"})]}),"\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u652f\u6301\u534f\u8bae\u5347\u7ea7\u548c\u5411\u540e\u517c\u5bb9"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u6d88\u606f\u7c7b\u578b (Message Type)"})," - 1\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsxs)(r.li,{children:["\u8bf7\u6c42\u6d88\u606f\uff1a",(0,i.jsx)(r.code,{children:"0x01"})]}),"\n",(0,i.jsxs)(r.li,{children:["\u54cd\u5e94\u6d88\u606f\uff1a",(0,i.jsx)(r.code,{children:"0x02"})]}),"\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u533a\u5206\u8bf7\u6c42\u548c\u54cd\u5e94\u6d88\u606f"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u538b\u7f29\u7c7b\u578b (Compress Type)"})," - 1\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsxs)(r.li,{children:["\u4e0d\u538b\u7f29\uff1a",(0,i.jsx)(r.code,{children:"0x00"})]}),"\n",(0,i.jsxs)(r.li,{children:["Gzip\u538b\u7f29\uff1a",(0,i.jsx)(r.code,{children:"0x01"})]}),"\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u6307\u793a\u6d88\u606f\u4f53\u662f\u5426\u538b\u7f29\u53ca\u4f7f\u7528\u7684\u538b\u7f29\u7b97\u6cd5"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u5e8f\u5217\u5316\u7c7b\u578b (Serialization Type)"})," - 1\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsxs)(r.li,{children:["JSON\u5e8f\u5217\u5316\uff1a",(0,i.jsx)(r.code,{children:"0x01"})]}),"\n",(0,i.jsxs)(r.li,{children:["Protobuf\u5e8f\u5217\u5316\uff1a",(0,i.jsx)(r.code,{children:"0x02"})]}),"\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u6307\u5b9a\u5143\u6570\u636e\u548c\u6d88\u606f\u4f53\u7684\u5e8f\u5217\u5316\u65b9\u5f0f"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u6d88\u606fID (Message ID)"})," - 8\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u552f\u4e00\u6807\u8bc6\u4e00\u4e2aRPC\u8bf7\u6c42\uff0c\u7528\u4e8e\u8bf7\u6c42\u548c\u54cd\u5e94\u7684\u914d\u5bf9\uff0c\u652f\u6301\u5f02\u6b65\u8c03\u7528\u548c\u591a\u8def\u590d\u7528"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u5143\u6570\u636e\u957f\u5ea6 (Metadata Size)"})," - 4\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u6307\u5b9a\u5143\u6570\u636e\u90e8\u5206\u7684\u5b57\u8282\u957f\u5ea6"}),"\n"]}),"\n"]}),"\n",(0,i.jsxs)(r.li,{children:["\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.strong,{children:"\u6d88\u606f\u4f53\u957f\u5ea6 (Payload Size)"})," - 4\u5b57\u8282"]}),"\n",(0,i.jsxs)(r.ul,{children:["\n",(0,i.jsx)(r.li,{children:"\u4f5c\u7528\uff1a\u6307\u5b9a\u6d88\u606f\u4f53\u90e8\u5206\u7684\u5b57\u8282\u957f\u5ea6"}),"\n"]}),"\n"]}),"\n"]}),"\n",(0,i.jsx)(r.h2,{id:"\u6d88\u606f\u7f16\u89e3\u7801",children:"\u6d88\u606f\u7f16\u89e3\u7801"}),"\n",(0,i.jsx)(r.p,{children:"fyerrpc\u4f7f\u7528Protocol\u63a5\u53e3\u5b9a\u4e49\u4e86\u6d88\u606f\u7684\u7f16\u7801\u548c\u89e3\u7801\u884c\u4e3a\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"type Protocol interface {\r\n    EncodeMessage(message *Message, writer io.Writer) error\r\n    DecodeMessage(reader io.Reader) (*Message, error)\r\n}\n"})}),"\n",(0,i.jsx)(r.h3,{id:"\u9ed8\u8ba4\u534f\u8bae\u5b9e\u73b0",children:"\u9ed8\u8ba4\u534f\u8bae\u5b9e\u73b0"}),"\n",(0,i.jsxs)(r.p,{children:[(0,i.jsx)(r.code,{children:"DefaultProtocol"}),"\u662f\u6846\u67b6\u63d0\u4f9b\u7684\u6807\u51c6\u5b9e\u73b0\uff0c\u5b83\u6309\u7167\u4e8c\u8fdb\u5236\u683c\u5f0f\u7f16\u7801\u548c\u89e3\u7801\u6d88\u606f\uff1a"]}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"// DefaultProtocol \u9ed8\u8ba4\u534f\u8bae\u5b9e\u73b0\r\ntype DefaultProtocol struct{}\r\n\r\n// EncodeMessage \u7f16\u7801\u6d88\u606f\r\nfunc (p *DefaultProtocol) EncodeMessage(message *Message, writer io.Writer) error {\r\n    // \u5199\u5165\u5934\u90e8\u5404\u4e2a\u5b57\u6bb5\r\n    if err := binary.Write(writer, binary.BigEndian, message.Header.MagicNumber); err != nil {\r\n        return err\r\n    }\r\n    // ... \u5199\u5165\u5176\u4ed6\u5934\u90e8\u5b57\u6bb5 ...\r\n\r\n    // \u5e8f\u5217\u5316\u5143\u6570\u636e\r\n    var metadataBytes []byte\r\n    var err error\r\n    if message.Metadata != nil {\r\n        codec := GetCodecByType(message.Header.SerializationType)\r\n        if codec == nil {\r\n            return ErrUnsupportedSerializer\r\n        }\r\n\r\n        metadataBytes, err = codec.Encode(message.Metadata)\r\n        if err != nil {\r\n            return err\r\n        }\r\n    }\r\n\r\n    // \u5199\u5165\u5143\u6570\u636e\u957f\u5ea6\r\n    message.Header.MetadataSize = uint32(len(metadataBytes))\r\n    if err := binary.Write(writer, binary.BigEndian, message.Header.MetadataSize); err != nil {\r\n        return err\r\n    }\r\n\r\n    // \u5199\u5165\u6d88\u606f\u4f53\u957f\u5ea6\r\n    message.Header.PayloadSize = uint32(len(message.Payload))\r\n    if err := binary.Write(writer, binary.BigEndian, message.Header.PayloadSize); err != nil {\r\n        return err\r\n    }\r\n\r\n    // \u5199\u5165\u5143\u6570\u636e\r\n    if len(metadataBytes) > 0 {\r\n        if _, err := writer.Write(metadataBytes); err != nil {\r\n            return err\r\n        }\r\n    }\r\n\r\n    // \u5199\u5165\u6d88\u606f\u4f53\r\n    if len(message.Payload) > 0 {\r\n        if _, err := writer.Write(message.Payload); err != nil {\r\n            return err\r\n        }\r\n    }\r\n\r\n    return nil\r\n}\n"})}),"\n",(0,i.jsx)(r.p,{children:"\u89e3\u7801\u8fc7\u7a0b\u662f\u7f16\u7801\u7684\u9006\u8fc7\u7a0b\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"// DecodeMessage \u89e3\u7801\u6d88\u606f\r\nfunc (p *DefaultProtocol) DecodeMessage(reader io.Reader) (*Message, error) {\r\n    message := &Message{\r\n        Header: Header{},\r\n    }\r\n\r\n    // \u8bfb\u53d6\u5934\u90e8\u5404\u4e2a\u5b57\u6bb5\r\n    if err := binary.Read(reader, binary.BigEndian, &message.Header.MagicNumber); err != nil {\r\n        return nil, err\r\n    }\r\n    if message.Header.MagicNumber != MagicNumber {\r\n        return nil, ErrInvalidMagic\r\n    }\r\n\r\n    // ... \u8bfb\u53d6\u5176\u4ed6\u5934\u90e8\u5b57\u6bb5 ...\r\n\r\n    // \u8bfb\u53d6\u5143\u6570\u636e\r\n    if message.Header.MetadataSize > 0 {\r\n        metadataBytes := make([]byte, message.Header.MetadataSize)\r\n        if _, err := io.ReadFull(reader, metadataBytes); err != nil {\r\n            return nil, err\r\n        }\r\n\r\n        codec := GetCodecByType(message.Header.SerializationType)\r\n        if codec == nil {\r\n            return nil, ErrUnsupportedSerializer\r\n        }\r\n\r\n        message.Metadata = &Metadata{}\r\n        if err := codec.Decode(metadataBytes, message.Metadata); err != nil {\r\n            return nil, err\r\n        }\r\n    }\r\n\r\n    // \u8bfb\u53d6\u6d88\u606f\u4f53\r\n    if message.Header.PayloadSize > 0 {\r\n        payload := make([]byte, message.Header.PayloadSize)\r\n        if _, err := io.ReadFull(reader, payload); err != nil {\r\n            return nil, err\r\n        }\r\n        message.Payload = payload\r\n    }\r\n\r\n    return message, nil\r\n}\n"})}),"\n",(0,i.jsx)(r.h2,{id:"\u5e8f\u5217\u5316",children:"\u5e8f\u5217\u5316"}),"\n",(0,i.jsxs)(r.p,{children:["fyerrpc\u652f\u6301\u591a\u79cd\u5e8f\u5217\u5316\u65b9\u5f0f\uff0c\u901a\u8fc7\u534f\u8bae\u5934\u7684",(0,i.jsx)(r.code,{children:"SerializationType"}),"\u5b57\u6bb5\u6307\u5b9a\u3002\u5e8f\u5217\u5316\u903b\u8f91\u7531",(0,i.jsx)(r.code,{children:"codec.Codec"}),"\u63a5\u53e3\u5b9a\u4e49\uff1a"]}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"type Codec interface {\r\n    // Encode \u5c06\u5bf9\u8c61\u5e8f\u5217\u5316\u4e3a\u5b57\u8282\u6570\u7ec4\r\n    Encode(v interface{}) ([]byte, error)\r\n\r\n    // Decode \u5c06\u5b57\u8282\u6570\u7ec4\u53cd\u5e8f\u5217\u5316\u4e3a\u5bf9\u8c61\r\n    Decode(data []byte, v interface{}) error\r\n\r\n    // Name \u8fd4\u56de\u7f16\u89e3\u7801\u5668\u7684\u540d\u79f0\r\n    Name() string\r\n}\n"})}),"\n",(0,i.jsx)(r.h3,{id:"\u5185\u7f6e\u5e8f\u5217\u5316\u5b9e\u73b0",children:"\u5185\u7f6e\u5e8f\u5217\u5316\u5b9e\u73b0"}),"\n",(0,i.jsx)(r.h4,{id:"json-\u5e8f\u5217\u5316",children:"JSON \u5e8f\u5217\u5316"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:'// JsonCodec \u5b9e\u73b0\u4e86 Codec \u63a5\u53e3\r\ntype JsonCodec struct{}\r\n\r\n// Encode \u5c06\u5bf9\u8c61\u5e8f\u5217\u5316\u4e3a JSON \u5b57\u8282\u6570\u7ec4\r\nfunc (c *JsonCodec) Encode(v interface{}) ([]byte, error) {\r\n    return json.Marshal(v)\r\n}\r\n\r\n// Decode \u5c06 JSON \u5b57\u8282\u6570\u7ec4\u53cd\u5e8f\u5217\u5316\u4e3a\u5bf9\u8c61\r\nfunc (c *JsonCodec) Decode(data []byte, v interface{}) error {\r\n    return json.Unmarshal(data, v)\r\n}\r\n\r\n// Name \u8fd4\u56de\u7f16\u89e3\u7801\u5668\u7684\u540d\u79f0\r\nfunc (c *JsonCodec) Name() string {\r\n    return "json"\r\n}\n'})}),"\n",(0,i.jsx)(r.h4,{id:"protobuf-\u5e8f\u5217\u5316",children:"Protobuf \u5e8f\u5217\u5316"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:'// ProtobufCodec \u5b9e\u73b0\u4e86 Codec \u63a5\u53e3\r\ntype ProtobufCodec struct{}\r\n\r\n// Encode \u5c06\u5bf9\u8c61\u5e8f\u5217\u5316\u4e3a Protobuf \u5b57\u8282\u6570\u7ec4\r\nfunc (c *ProtobufCodec) Encode(v interface{}) ([]byte, error) {\r\n    // \u7c7b\u578b\u65ad\u8a00\u786e\u4fddv\u662fproto.Message\u7c7b\u578b\r\n    if pm, ok := v.(proto.Message); ok {\r\n        return proto.Marshal(pm)\r\n    }\r\n    return nil, ErrInvalidMessage\r\n}\r\n\r\n// Decode \u5c06 Protobuf \u5b57\u8282\u6570\u7ec4\u53cd\u5e8f\u5217\u5316\u4e3a\u5bf9\u8c61\r\nfunc (c *ProtobufCodec) Decode(data []byte, v interface{}) error {\r\n    // \u7c7b\u578b\u65ad\u8a00\u786e\u4fddv\u662fproto.Message\u7c7b\u578b\r\n    if pm, ok := v.(proto.Message); ok {\r\n        return proto.Unmarshal(data, pm)\r\n    }\r\n    return ErrInvalidMessage\r\n}\r\n\r\n// Name \u8fd4\u56de\u7f16\u89e3\u7801\u5668\u7684\u540d\u79f0\r\nfunc (c *ProtobufCodec) Name() string {\r\n    return "protobuf"\r\n}\n'})}),"\n",(0,i.jsx)(r.h2,{id:"\u81ea\u5b9a\u4e49\u534f\u8bae",children:"\u81ea\u5b9a\u4e49\u534f\u8bae"}),"\n",(0,i.jsx)(r.p,{children:"fyerrpc\u652f\u6301\u81ea\u5b9a\u4e49\u534f\u8bae\uff0c\u60a8\u53ef\u4ee5\u6269\u5c55\u6216\u5b8c\u5168\u66ff\u6362\u9ed8\u8ba4\u7684\u534f\u8bae\u5b9e\u73b0\u3002"}),"\n",(0,i.jsx)(r.h3,{id:"\u6269\u5c55\u9ed8\u8ba4\u534f\u8bae",children:"\u6269\u5c55\u9ed8\u8ba4\u534f\u8bae"}),"\n",(0,i.jsx)(r.p,{children:"\u6269\u5c55\u9ed8\u8ba4\u534f\u8bae\u6700\u7b80\u5355\u7684\u65b9\u5f0f\u662f\u5728\u73b0\u6709\u534f\u8bae\u57fa\u7840\u4e0a\u6dfb\u52a0\u529f\u80fd\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"// EnhancedProtocol \u6269\u5c55\u9ed8\u8ba4\u534f\u8bae\uff0c\u6dfb\u52a0\u52a0\u5bc6\u529f\u80fd\r\ntype EnhancedProtocol struct {\r\n    DefaultProtocol\r\n    encryptionKey []byte\r\n}\r\n\r\n// EncodeMessage \u91cd\u5199\u7f16\u7801\u65b9\u6cd5\uff0c\u6dfb\u52a0\u52a0\u5bc6\r\nfunc (p *EnhancedProtocol) EncodeMessage(message *Message, writer io.Writer) error {\r\n    // \u52a0\u5bc6\u6d88\u606f\u4f53\r\n    if len(message.Payload) > 0 {\r\n        encrypted, err := encrypt(message.Payload, p.encryptionKey)\r\n        if err != nil {\r\n            return err\r\n        }\r\n        message.Payload = encrypted\r\n    }\r\n    \r\n    // \u8c03\u7528\u9ed8\u8ba4\u5b9e\u73b0\u5b8c\u6210\u7f16\u7801\r\n    return p.DefaultProtocol.EncodeMessage(message, writer)\r\n}\r\n\r\n// DecodeMessage \u91cd\u5199\u89e3\u7801\u65b9\u6cd5\uff0c\u6dfb\u52a0\u89e3\u5bc6\r\nfunc (p *EnhancedProtocol) DecodeMessage(reader io.Reader) (*Message, error) {\r\n    // \u5148\u4f7f\u7528\u9ed8\u8ba4\u5b9e\u73b0\u89e3\u7801\r\n    message, err := p.DefaultProtocol.DecodeMessage(reader)\r\n    if err != nil {\r\n        return nil, err\r\n    }\r\n    \r\n    // \u89e3\u5bc6\u6d88\u606f\u4f53\r\n    if len(message.Payload) > 0 {\r\n        decrypted, err := decrypt(message.Payload, p.encryptionKey)\r\n        if err != nil {\r\n            return nil, err\r\n        }\r\n        message.Payload = decrypted\r\n    }\r\n    \r\n    return message, nil\r\n}\r\n\r\n// \u521b\u5efa\u52a0\u5bc6\u534f\u8bae\u5b9e\u4f8b\r\nfunc NewEncryptedProtocol(key []byte) *EnhancedProtocol {\r\n    return &EnhancedProtocol{\r\n        encryptionKey: key,\r\n    }\r\n}\n"})}),"\n",(0,i.jsx)(r.h3,{id:"\u5b9e\u73b0\u5168\u65b0\u534f\u8bae",children:"\u5b9e\u73b0\u5168\u65b0\u534f\u8bae"}),"\n",(0,i.jsx)(r.p,{children:"\u5982\u679c\u9ed8\u8ba4\u534f\u8bae\u4e0d\u6ee1\u8db3\u9700\u6c42\uff0c\u60a8\u53ef\u4ee5\u5b9e\u73b0\u5b8c\u5168\u4e0d\u540c\u7684\u534f\u8bae\u683c\u5f0f\uff1a"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:"// CompactProtocol \u5b9e\u73b0\u66f4\u7d27\u51d1\u7684\u534f\u8bae\u683c\u5f0f\r\ntype CompactProtocol struct{}\r\n\r\n// EncodeMessage \u4f7f\u7528\u7d27\u51d1\u683c\u5f0f\u7f16\u7801\u6d88\u606f\r\nfunc (p *CompactProtocol) EncodeMessage(message *Message, writer io.Writer) error {\r\n    // \u5b9e\u73b0\u7d27\u51d1\u7f16\u7801\u903b\u8f91\r\n    // \u4f8b\u5982\uff1a\u4f7f\u7528\u53d8\u957f\u6574\u6570\u7f16\u7801\u3001\u4f4d\u538b\u7f29\u7b49\u6280\u672f\u51cf\u5c11\u534f\u8bae\u5f00\u9500\r\n    \r\n    // \u793a\u4f8b\uff1a\u4f7f\u7528varint\u7f16\u7801\u6d88\u606fID\r\n    var buf [10]byte\r\n    n := binary.PutUvarint(buf[:], message.Header.MessageID)\r\n    if _, err := writer.Write(buf[:n]); err != nil {\r\n        return err\r\n    }\r\n    \r\n    // ... \u7f16\u7801\u5176\u4ed6\u5b57\u6bb5 ...\r\n    \r\n    return nil\r\n}\r\n\r\n// DecodeMessage \u89e3\u7801\u7d27\u51d1\u683c\u5f0f\u6d88\u606f\r\nfunc (p *CompactProtocol) DecodeMessage(reader io.Reader) (*Message, error) {\r\n    message := &Message{\r\n        Header: Header{},\r\n    }\r\n    \r\n    // \u5b9e\u73b0\u7d27\u51d1\u89e3\u7801\u903b\u8f91\r\n    // \u793a\u4f8b\uff1a\u4f7f\u7528varint\u89e3\u7801\u6d88\u606fID\r\n    messageID, err := binary.ReadUvarint(reader.(io.ByteReader))\r\n    if err != nil {\r\n        return nil, err\r\n    }\r\n    message.Header.MessageID = messageID\r\n    \r\n    // ... \u89e3\u7801\u5176\u4ed6\u5b57\u6bb5 ...\r\n    \r\n    return message, nil\r\n}\n"})}),"\n",(0,i.jsx)(r.h3,{id:"\u6ce8\u518c\u81ea\u5b9a\u4e49\u5e8f\u5217\u5316\u5668",children:"\u6ce8\u518c\u81ea\u5b9a\u4e49\u5e8f\u5217\u5316\u5668"}),"\n",(0,i.jsxs)(r.p,{children:["\u8981\u6dfb\u52a0\u65b0\u7684\u5e8f\u5217\u5316\u65b9\u5f0f\uff0c\u5b9e\u73b0\u5e76\u6ce8\u518c",(0,i.jsx)(r.code,{children:"Codec"}),"\u63a5\u53e3\uff1a"]}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:'// MessagePack\u5e8f\u5217\u5316\u5668\u793a\u4f8b\r\ntype MsgpackCodec struct{}\r\n\r\nfunc (c *MsgpackCodec) Encode(v interface{}) ([]byte, error) {\r\n    return msgpack.Marshal(v)\r\n}\r\n\r\nfunc (c *MsgpackCodec) Decode(data []byte, v interface{}) error {\r\n    return msgpack.Unmarshal(data, v)\r\n}\r\n\r\nfunc (c *MsgpackCodec) Name() string {\r\n    return "msgpack"\r\n}\r\n\r\n// \u5b9a\u4e49\u65b0\u7684\u5e8f\u5217\u5316\u7c7b\u578b\r\nconst (\r\n    SerializationTypeMsgpack = uint8(0x03) // MessagePack\u5e8f\u5217\u5316\r\n)\r\n\r\n// \u6ce8\u518c\u65b0\u7684\u5e8f\u5217\u5316\u5668\r\nfunc init() {\r\n    codec.RegisterCodec(codec.Type(2), &MsgpackCodec{})\r\n}\r\n\r\n// \u6269\u5c55GetCodecByType\u51fd\u6570\r\nfunc GetCodecByType(serializationType uint8) codec.Codec {\r\n    switch serializationType {\r\n    case SerializationTypeJSON:\r\n        return codec.GetCodec(codec.JSON)\r\n    case SerializationTypeProtobuf:\r\n        return codec.GetCodec(codec.Protobuf)\r\n    case SerializationTypeMsgpack:\r\n        return codec.GetCodec(codec.Type(2))\r\n    default:\r\n        return nil\r\n    }\r\n}\n'})}),"\n",(0,i.jsx)(r.h2,{id:"\u534f\u8bae\u4f7f\u7528\u793a\u4f8b",children:"\u534f\u8bae\u4f7f\u7528\u793a\u4f8b"}),"\n",(0,i.jsx)(r.h3,{id:"\u57fa\u672c\u4f7f\u7528",children:"\u57fa\u672c\u4f7f\u7528"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:'// \u521b\u5efa\u534f\u8bae\u5b9e\u4f8b\r\nprotocol := &protocol.DefaultProtocol{}\r\n\r\n// \u521b\u5efa\u8bf7\u6c42\u6d88\u606f\r\nmessage := &protocol.Message{\r\n    Header: protocol.Header{\r\n        MagicNumber:       protocol.MagicNumber,\r\n        Version:           1,\r\n        MessageType:       protocol.TypeRequest,\r\n        CompressType:      protocol.CompressTypeNone,\r\n        SerializationType: protocol.SerializationTypeJSON,\r\n        MessageID:         1234567890,\r\n    },\r\n    Metadata: &protocol.Metadata{\r\n        ServiceName: "UserService",\r\n        MethodName:  "GetUser",\r\n        Extra: map[string]string{\r\n            "trace_id": "abc123",\r\n            "user_id":  "1001",\r\n        },\r\n    },\r\n    Payload: []byte(`{"id": 1}`),\r\n}\r\n\r\n// \u7f16\u7801\u6d88\u606f\r\nbuf := new(bytes.Buffer)\r\nerr := protocol.EncodeMessage(message, buf)\r\nif err != nil {\r\n    log.Fatalf("\u7f16\u7801\u6d88\u606f\u5931\u8d25: %v", err)\r\n}\r\n\r\n// \u89e3\u7801\u6d88\u606f\r\ndecoded, err := protocol.DecodeMessage(buf)\r\nif err != nil {\r\n    log.Fatalf("\u89e3\u7801\u6d88\u606f\u5931\u8d25: %v", err)\r\n}\r\n\r\nfmt.Printf("\u670d\u52a1\u540d: %s, \u65b9\u6cd5\u540d: %s\\n", \r\n    decoded.Metadata.ServiceName, \r\n    decoded.Metadata.MethodName)\n'})}),"\n",(0,i.jsx)(r.h3,{id:"\u4f7f\u7528\u81ea\u5b9a\u4e49\u534f\u8bae",children:"\u4f7f\u7528\u81ea\u5b9a\u4e49\u534f\u8bae"}),"\n",(0,i.jsx)(r.pre,{children:(0,i.jsx)(r.code,{className:"language-go",children:'// \u521b\u5efa\u52a0\u5bc6\u534f\u8bae\u5b9e\u4f8b\r\nencryptedProtocol := NewEncryptedProtocol([]byte("secret-key-12345"))\r\n\r\n// \u7f16\u7801\u548c\u89e3\u7801\u6d88\u606f\r\nbuf := new(bytes.Buffer)\r\nerr := encryptedProtocol.EncodeMessage(message, buf)\r\nif err != nil {\r\n    log.Fatalf("\u52a0\u5bc6\u7f16\u7801\u6d88\u606f\u5931\u8d25: %v", err)\r\n}\r\n\r\ndecoded, err := encryptedProtocol.DecodeMessage(buf)\r\nif err != nil {\r\n    log.Fatalf("\u89e3\u5bc6\u89e3\u7801\u6d88\u606f\u5931\u8d25: %v", err)\r\n}\n'})}),"\n",(0,i.jsx)(r.h2,{id:"\u534f\u8bae\u8bbe\u8ba1\u8003\u91cf",children:"\u534f\u8bae\u8bbe\u8ba1\u8003\u91cf"}),"\n",(0,i.jsx)(r.p,{children:"fyerrpc\u534f\u8bae\u8bbe\u8ba1\u8fc7\u7a0b\u4e2d\u8003\u8651\u4e86\u591a\u79cd\u56e0\u7d20\uff1a"}),"\n",(0,i.jsx)(r.h3,{id:"\u6027\u80fd\u4f18\u5316",children:"\u6027\u80fd\u4f18\u5316"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u7d27\u51d1\u7684\u4e8c\u8fdb\u5236\u683c\u5f0f"}),"\uff1a\u6bd4\u6587\u672c\u683c\u5f0f\uff08\u5982JSON-RPC\uff09\u66f4\u9ad8\u6548"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u56fa\u5b9a\u957f\u5ea6\u5934\u90e8"}),"\uff1a\u5feb\u901f\u89e3\u6790\uff0c\u65e0\u9700\u626b\u63cf\u5206\u9694\u7b26"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u5b57\u8282\u5e8f\u4e00\u81f4\u6027"}),"\uff1a\u7edf\u4e00\u4f7f\u7528\u7f51\u7edc\u5b57\u8282\u5e8f\uff08\u5927\u7aef\u5e8f\uff09\uff0c\u907f\u514d\u8de8\u5e73\u53f0\u95ee\u9898"]}),"\n"]}),"\n",(0,i.jsx)(r.h3,{id:"\u53ef\u6269\u5c55\u6027",children:"\u53ef\u6269\u5c55\u6027"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u7248\u672c\u5b57\u6bb5"}),"\uff1a\u652f\u6301\u534f\u8bae\u6f14\u8fdb\u548c\u5411\u540e\u517c\u5bb9"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u53ef\u6269\u5c55\u7684\u5143\u6570\u636e"}),"\uff1a\u901a\u8fc7Extra\u5b57\u6bb5\u652f\u6301\u81ea\u5b9a\u4e49\u5c5e\u6027"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u53ef\u63d2\u62d4\u7684\u5e8f\u5217\u5316\u65b9\u5f0f"}),"\uff1a\u652f\u6301\u4e0d\u540c\u573a\u666f\u9009\u62e9\u6700\u9002\u5408\u7684\u5e8f\u5217\u5316\u65b9\u6848"]}),"\n"]}),"\n",(0,i.jsx)(r.h3,{id:"\u53ef\u8c03\u8bd5\u6027",children:"\u53ef\u8c03\u8bd5\u6027"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u9b54\u6570\u6821\u9a8c"}),"\uff1a\u5feb\u901f\u8bc6\u522b\u65e0\u6548\u6d88\u606f"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u6e05\u6670\u7684\u5b57\u6bb5\u5212\u5206"}),"\uff1a\u4fbf\u4e8e\u95ee\u9898\u6392\u67e5"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u53cb\u597d\u7684\u9519\u8bef\u63d0\u793a"}),"\uff1a\u8be6\u7ec6\u9519\u8bef\u4fe1\u606f\u5e2e\u52a9\u5b9a\u4f4d\u95ee\u9898"]}),"\n"]}),"\n",(0,i.jsx)(r.h2,{id:"\u6700\u4f73\u5b9e\u8df5",children:"\u6700\u4f73\u5b9e\u8df5"}),"\n",(0,i.jsx)(r.h3,{id:"\u534f\u8bae\u9009\u62e9",children:"\u534f\u8bae\u9009\u62e9"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u5185\u90e8\u7cfb\u7edf\u901a\u4fe1"}),"\uff1a\u4f7f\u7528\u9ed8\u8ba4\u534f\u8bae\u548cProtobuf\u5e8f\u5217\u5316\uff0c\u517c\u987e\u6027\u80fd\u548c\u7c7b\u578b\u5b89\u5168"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u5bf9\u5916API"}),"\uff1a\u8003\u8651\u4f7f\u7528HTTP+JSON\u7684RESTful\u63a5\u53e3\uff0c\u517c\u5bb9\u6027\u66f4\u597d"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u9ad8\u5b89\u5168\u9700\u6c42"}),"\uff1a\u4f7f\u7528\u6269\u5c55\u7684\u52a0\u5bc6\u534f\u8bae\u4fdd\u62a4\u654f\u611f\u6570\u636e"]}),"\n"]}),"\n",(0,i.jsx)(r.h3,{id:"\u534f\u8bae\u6269\u5c55",children:"\u534f\u8bae\u6269\u5c55"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u4fdd\u7559\u5411\u540e\u517c\u5bb9\u6027"}),"\uff1a\u534f\u8bae\u5347\u7ea7\u65f6\u4fdd\u8bc1\u65e7\u7248\u672c\u5ba2\u6237\u7aef\u4ecd\u80fd\u6b63\u5e38\u5de5\u4f5c"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u76d1\u63a7\u534f\u8bae\u9519\u8bef"}),"\uff1a\u6536\u96c6\u548c\u5206\u6790\u534f\u8bae\u89e3\u6790\u9519\u8bef\uff0c\u53ca\u65f6\u53d1\u73b0\u95ee\u9898"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u6027\u80fd\u57fa\u51c6\u6d4b\u8bd5"}),"\uff1a\u5b9a\u671f\u6d4b\u8bd5\u534f\u8bae\u6027\u80fd\uff0c\u7279\u522b\u662f\u5728\u4fee\u6539\u540e"]}),"\n"]}),"\n",(0,i.jsx)(r.h3,{id:"\u5e8f\u5217\u5316\u9009\u62e9",children:"\u5e8f\u5217\u5316\u9009\u62e9"}),"\n",(0,i.jsxs)(r.ol,{children:["\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u5f00\u53d1\u73af\u5883"}),"\uff1a\u63a8\u8350JSON\u5e8f\u5217\u5316\uff0c\u65b9\u4fbf\u8c03\u8bd5\u548c\u65e5\u5fd7\u5206\u6790"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u751f\u4ea7\u73af\u5883"}),"\uff1a\u63a8\u8350Protobuf\u5e8f\u5217\u5316\uff0c\u6027\u80fd\u66f4\u597d\uff0c\u6570\u636e\u66f4\u7d27\u51d1"]}),"\n",(0,i.jsxs)(r.li,{children:[(0,i.jsx)(r.strong,{children:"\u7279\u6b8a\u573a\u666f"}),"\uff1a\u6839\u636e\u9700\u6c42\u9009\u62e9\u6216\u5b9e\u73b0\u81ea\u5b9a\u4e49\u5e8f\u5217\u5316\u5668"]}),"\n"]}),"\n",(0,i.jsx)(r.p,{children:"\u901a\u8fc7\u672c\u6587\u6863\uff0c\u60a8\u5df2\u7ecf\u4e86\u89e3\u4e86fyerrpc\u534f\u8bae\u7684\u8be6\u7ec6\u8bbe\u8ba1\u3001\u4f7f\u7528\u65b9\u6cd5\u548c\u6269\u5c55\u673a\u5236\u3002\u8fd9\u4e9b\u77e5\u8bc6\u5c06\u5e2e\u52a9\u60a8\u66f4\u597d\u5730\u4f7f\u7528fyerrpc\u6846\u67b6\uff0c\u4e5f\u4e3a\u81ea\u5b9a\u4e49\u548c\u6269\u5c55\u534f\u8bae\u63d0\u4f9b\u4e86\u6307\u5bfc\u3002"})]})}function h(e={}){const{wrapper:r}={...(0,a.R)(),...e.components};return r?(0,i.jsx)(r,{...e,children:(0,i.jsx)(t,{...e})}):t(e)}},8453:(e,r,n)=>{n.d(r,{R:()=>c,x:()=>l});var s=n(6540);const i={},a=s.createContext(i);function c(e){const r=s.useContext(a);return s.useMemo((function(){return"function"==typeof e?e(r):{...r,...e}}),[r,e])}function l(e){let r;return r=e.disableParentContext?"function"==typeof e.components?e.components(i):e.components||i:c(e.components),s.createElement(a.Provider,{value:r},e.children)}}}]);