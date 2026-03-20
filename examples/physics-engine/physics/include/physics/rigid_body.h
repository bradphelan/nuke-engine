#pragma once
#include <mathcore/vec3.h>
namespace physics {
struct RigidBody {
    mathcore::Vec3 position;
    mathcore::Vec3 velocity;
    mathcore::Vec3 acceleration;
    float mass;
    RigidBody(float mass = 1.0f);
    void apply_force(const mathcore::Vec3& force);
    void step(float dt);
};
}
