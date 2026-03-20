#include <renderer/renderer.h>
#include <cstdio>

namespace renderer {

Renderer::Renderer() : frame_count_(0) {}

void Renderer::draw_point(const mathcore::Vec3& pos, const Color& col) {
    printf("  draw (%.2f, %.2f, %.2f) color (%.1f, %.1f, %.1f)\n",
           pos.x, pos.y, pos.z, col.r, col.g, col.b);
}

void Renderer::present() {
    frame_count_++;
    printf("  [frame %d presented]\n", frame_count_);
}

}
