#version 450
#extension GL_EXT_nonuniform_qualifier : require

// must use mat4 -- mat3 alignment issues are horrible.
// each mat4 = 64 bytes, so full 128 byte total, but only using mat3.
// pack the tex, layer indexes into [3][0-1] of mvp,
// and the fill color into [3][0-3] of uvp
layout(push_constant) uniform Mtxs {
	mat4 mvp;
	mat4 uvp;
};

layout(set = 0, binding = 0) uniform sampler2DArray Tex[];

layout(location = 0) in vector2 uv;
layout(location = 0) out vector4 outputColor;

void main() {
	int idx = int(mvp[3][0]);
	int layer = int(mvp[3][1]);
	outputColor = texture(Tex[idx], vector3(uv,layer));
}

