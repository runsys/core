#version 450 core

layout(push_constant) uniform TexIdxUni {
    int TexIdx;
};

layout(set = 1, binding = 0) uniform sampler2D TexSampler[3];

layout(location = 0) in vec3 FragColor;
layout(location = 1) in vec2 FragTexCoord;

layout(location = 0) out vec4 OutColor;

void main() {
    OutColor = texture(TexSampler[TexIdx], FragTexCoord);
    // OutColor = vec4(FragColor, 1.0);
}

