import React from 'react';
import { hot } from 'react-hot-loader';
import './App.css';

// component import

export class App extends React.Component {
  constructor(props){
    // super
    super(props)

    // States
    this.state = {
      preview: 0,
      program: 0,
      connected: false
    }

    this.setting = {
      input:1,
      meIndex:1,
    }


    this.pluginAction = null
    this.uuid = ''
    this.context = ""
  }

  componentDidMount(){
      if (window.$SD) {
        window.$SD.on('connected', (jsonObj)=> {
          console.log("connected", jsonObj)
          this.uuid = jsonObj['uuid'];
          if (jsonObj.hasOwnProperty('actionInfo')) {
            this.pluginAction = jsonObj.actionInfo['action'];
            this.context = jsonObj.actionInfo['context'];

            if (jsonObj.actionInfo.payload.hasOwnProperty("settings")){
            }
          }
          console.log("current state:",this.state)
        });

        window.$SD.on("sendToPropertyInspector", (jsonObj) => {
          console.log("sendToPropertyInspector", jsonObj)
          if(!jsonObj.payload){
            return
          }
          if(jsonObj.event === "sendToPropertyInspector"){
            if(Array.isArray(jsonObj.payload.inputs)){
              if(this.state.inputs.length !== jsonObj.payload.inputs.length){
                this.setState({inputs:jsonObj.payload.inputs})
              }
            }
          }
          else if(jsonObj.event === "didReceiveSettings"){
            console.log("didReceiveSettings", jsonObj.payload)
          }
        })
        window.$SD.on("didReceiveGlobalSettings", (jsonObj) => {
          console.log("didReceiveGlobalSettings")
        })
      }
    }

  saveSettings(){
    // TODO
  }

  FunctionNameChange(funcName){
    this.setState({functionName:funcName}, ()=>{
      this.saveSettings()
    })
  }

  render(){
    return (
      <div className="App">
      {/* Wrapper starts from here... */}
        <div className="sdpi-wrapper">
        </div>
      </div>
    );
  }
}

export default hot(module)(App);
