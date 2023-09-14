package static

import (
	"bytes"
	"html/template"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

const htmlDeploymentTemplate = `<html lang="en"><head><meta http-equiv="refresh" content="10"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="theme-color" content="#000"><title>Loading - {{.Framework}}</title><style>
.tooltip {
    position: relative;
    display: inline-block;
    border-bottom: 1px dotted black;
  }
  
  .tooltip .tooltiptext {
    visibility: hidden;
    width: 120px;
    background-color: black;
    color: #fff;
    text-align: center;
    border-radius: 6px;
    padding: 5px 0;
    
    /* Position the tooltip */
    position: absolute;
    z-index: 1;
    bottom: 100%;
    left: 50%;
    margin-left: -60px;
  }
  
  .tooltip:hover .tooltiptext {
    visibility: visible;
  }
  html{font-size:62.5%;box-sizing:border-box;height:-webkit-fill-available}*,::after,::before{box-sizing:inherit}body{font-family:sf pro text,sf pro icons,helvetica neue,helvetica,arial,sans-serif;font-size:1.6rem;line-height:1.65;word-break:break-word;font-kerning:auto;font-variant:normal;-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale;text-rendering:optimizeLegibility;hyphens:auto;height:100vh;height:-webkit-fill-available;max-height:100vh;max-height:-webkit-fill-available;margin:0}::selection{background:#dbe6d2}::-moz-selection{background:#dbe6d2}a{cursor:pointer;color:#5e785f;text-decoration:none;transition:all .2s ease;border-bottom:1px solid #0000}a:hover{border-bottom:1px solid #5e785f}ul{padding:0;margin-left:1.5em;list-style-type:none}li{margin-bottom:10px}ul li:before{content:'\02013'}li:before{display:inline-block;color:#ccc;position:absolute;margin-left:-18px;transition:color .2s ease}code{font-family:Menlo,Monaco,Lucida Console,Liberation Mono,DejaVu Sans Mono,Bitstream Vera Sans Mono,Courier New,monospace,serif;font-size:.92em}code:after,code:before{content:'` + "`" + `'}.container{display:flex;justify-content:center;flex-direction:column;min-height:100%}main{max-width:80rem;padding:4rem 6rem;margin:auto}ul{margin-bottom:32px}.error-title{font-size:2rem;padding-left:22px;line-height:1.5;margin-bottom:24px}.error-title-guilty{border-left:2px solid #ed367f}.error-title-innocent{border-left:2px solid #59b89c}main p{color:#333}.devinfo-container{border:1px solid #ddd;border-radius:4px;padding:2rem;display:flex;flex-direction:column;margin-bottom:32px}.error-code{margin:0;font-size:1.6rem;color:#000;margin-bottom:1.6rem}.devinfo-line{color:#333}.devinfo-line code,code,li{color:#000}.devinfo-line:not(:last-child){margin-bottom:8px}.docs-link,.contact-link{font-weight:500}header,footer,footer a{display:flex;justify-content:center;align-items:center}header,footer{min-height:100px;height:100px}header{border-bottom:1px solid #eaeaea}header h1{font-size:1.8rem;margin:0;font-weight:500}header p{font-size:1.3rem;margin:0;font-weight:500}.header-item{display:flex;padding:0 2rem;margin:2rem 0;text-decoration:line-through;color:#999}.header-item.active{color:#ff0080;text-decoration:none}.header-item.first{border-right:1px solid #eaeaea}.header-item-content{display:flex;flex-direction:column}.header-item-icon{margin-right:1rem;margin-top:.6rem}footer{border-top:1px solid #eaeaea}footer a{color:#000}footer a:hover{border-bottom-color:#0000}footer svg{margin-left:.8rem}.note{padding:8pt 16pt;border-radius:5px;border:1px solid #0070f3;font-size:14px;line-height:1.8;color:#0070f3}@media(max-width:500px){.devinfo-container .devinfo-line code{margin-top:.4rem}.devinfo-container .devinfo-line:not(:last-child){margin-bottom:1.6rem}.devinfo-container{margin-bottom:0}header{flex-direction:column;height:auto;min-height:auto;align-items:flex-start}.header-item.first{border-right:none;margin-bottom:0}main{padding:1rem 2rem}body{font-size:1.4rem;line-height:1.55}footer{display:none}.note{margin-top:16px}}</style></head><body><div class="container"><main><p class="devinfo-container"><span class="error-code">{{.StatusString}}</span><span class="devinfo-line">Framework: <code>{{.Framework}}</code></span><span class="devinfo-line">Deployment: <code>{{.Deployment}}</code></span><span class="devinfo-line"><span class="tooltip"><span class="tooltiptext">Scheduling, ContainerCreating, Initializing, Running</span>Status</span>: <code>{{.InstanceStatus}}</code></span><span class="devinfo-line">The page will auto refresh once the request is completed. Kindly wait for the page to reload automatically. If the issue persists, please contact <a href="https://discord.gg/F4WnzqmeNj">modelz support team on discord</a> for assistance.</span></p></main></div></body></html>`

type htmlDeploymentStruct struct {
	Deployment     string
	Framework      string
	ID             string
	StatusString   string
	InstanceStatus string
}

func RenderDeploymentLoadingPage(framework, id, statusString, deployment string,
	instances []types.InferenceDeploymentInstance) (*bytes.Buffer, error) {
	tmpl, err := template.New("root").Parse(htmlDeploymentTemplate)
	if err != nil {
		return nil, err
	}

	data := htmlDeploymentStruct{
		Deployment:     deployment,
		Framework:      framework,
		ID:             id,
		StatusString:   statusString,
		InstanceStatus: "Scaling",
	}

	if len(instances) > 0 {
		data.InstanceStatus = string(instances[0].Status.Phase)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return nil, err
	}

	return &buffer, nil
}
