#version 450
#extension GL_EXT_nonuniform_qualifier : require

// must be <= 128 bytes -- contains all per-object data
layout(push_constant) uniform PushU {
	mat4 ModelMtx; // 64 bytes, [3][3] = TexPct.X
	vec4 Color; // 16
	vec4 ShinyBright; // 16 x = Shiny, y = Reflect, z = Bright, w = TexIndex
	vec4 Emissive; // 16 rgb, a = TexPct.Y
	vec4 TexRepeatOff; // 16 xy = Repeat, zw = Offset
};

layout(set = 0, binding = 0) uniform MtxsU {
    mat4 ViewMtx;
    mat4 PrjnMtx;
};

layout(location = 0) in vec4 Pos;
layout(location = 1) in vec3 Norm;
layout(location = 2) in vec3 CamDir;
// layout(location = 3) in vec2 TexCoord;
layout(location = 3) in vec4 VtxColor;

layout(location = 0) out vec4 outColor;

#include "phong_frag.frag"
			
void main() {
	float opacity = VtxColor.a;
	vec3 clr = SRGBToLinear(VtxColor.rgb);	// we need to undo gamma on incoming colors
	// vec3 clr = VtxColor.rgb;
	
	// Calculates the Ambient+Diffuse and Specular colors for this fragment using the Phong model.
	float Shiny = ShinyBright.x;
	float Reflect = 0; // ShinyBright.y;
	float Bright = ShinyBright.z;
	vec3 Specular = vec3(1,1,1);
	PhongModel(Pos, Norm, CamDir, clr, clr, Specular, Shiny, Reflect, Bright, opacity, outColor);
}

