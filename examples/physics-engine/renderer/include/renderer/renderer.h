#pragma once
#include <mathcore/vec3.h>
namespace renderer {
struct Color { float r, g, b; };

#ifdef RENDERER_EXPORTS
#define RENDERER_API __declspec(dllexport)
#else
#define RENDERER_API __declspec(dllimport)
#endif

class RENDERER_API Renderer {
public:
    Renderer();
    void draw_point(const mathcore::Vec3& pos, const Color& col);
    void present();
private:
    int frame_count_;
};
}
