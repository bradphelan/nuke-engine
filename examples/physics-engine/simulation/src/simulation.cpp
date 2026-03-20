#include <simulation/simulation.h>
#include <physics/rigid_body.h>
#include <renderer/renderer.h>

namespace simulation {

struct Simulation::Impl {
    physics::RigidBody bodies[16];
    int count = 0;
    renderer::Renderer renderer;
};

Simulation::Simulation() : impl_(new Impl()) {}

void Simulation::add_body(float mass, float x, float y, float z) {
    if (impl_->count >= 16) return;
    physics::RigidBody& body = impl_->bodies[impl_->count++];
    body = physics::RigidBody(mass);
    body.position = mathcore::Vec3(x, y, z);
}

void Simulation::step(float dt) {
    for (int i = 0; i < impl_->count; i++) {
        impl_->bodies[i].apply_force(mathcore::Vec3(0, -9.81f * impl_->bodies[i].mass, 0));
        impl_->bodies[i].step(dt);
    }
}

void Simulation::render() {
    renderer::Color white{1.0f, 1.0f, 1.0f};
    for (int i = 0; i < impl_->count; i++) {
        impl_->renderer.draw_point(impl_->bodies[i].position, white);
    }
    impl_->renderer.present();
}

int Simulation::body_count() const { return impl_->count; }

}
