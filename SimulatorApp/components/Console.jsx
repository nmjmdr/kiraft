import React from 'react';

export default class Console extends React.Component {

       componentDidMount() {
       	  this.updateCanvas();
       }

       updateCanvas() {
	   const ctx = this.refs.canvas.getContext('2d');	   
	   var dimensions = this.drawConsole(ctx);
	   var n = 20;

	   var center = {
	       x : dimensions.side/2,
	       y : dimensions.side/2
	   }

	   var rInner = 30;

	   var gap = rInner*2;
	   var r = (dimensions.side/2 - gap);
	  
           this.drawPath(ctx,r,center);	   

	   this.drawNodes(ctx,n,rInner,center,r);

	  
       }

       drawNodes(ctx,nodes,rInner,center,r) {       	
	   for (var i = 1; i <= nodes; ++i) {
  	       ctx.beginPath();
  	       var angle = i*2*Math.PI/nodes;
  	       var x = center.x + Math.cos(angle) * r;
	       var y = center.y + Math.sin(angle) * r;
      	       ctx.arc(x, y, rInner, 0, 360, false);
  	       ctx.fill();
	   }
       }

       drawPath(ctx,r,center) {
           ctx.beginPath();
	   ctx.arc(center.x, center.y, r, 0, 2 * Math.PI, false);
      	   ctx.stroke();
       }

       drawConsole(ctx) {
           var width = ctx.canvas.width;	
	   var filler = 20
           ctx.rect(filler,filler,width-filler,width-filler);
	   ctx.stroke();	  
	   var dimensions = {
	       side : (width+filler)	       
	   };

	   return dimensions;
       }


      render() {
      	       return (<div>
	       <canvas ref="canvas" width={700} height={700}/>
	       </div>);
      }

} 